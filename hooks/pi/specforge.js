// SpecForge enforcement adapter for the pi coding agent (pi.dev).
//
// This is a thin "translate pair" (see hooks/README.md): it maps pi's extension
// events to the SpecForge portable contract and back. ALL decision logic lives
// in the `sf` CLI (`sf hook --harness=generic`); this file only translates I/O.
// A new harness is a new adapter like this one, never new enforcement logic.
//
// Event mapping (pi → SpecForge):
//   tool_call (write|edit)   → pre_tool_use     (hard-deny a gate-skipping write)
//   session_start            → session_start    (context, injected next turn*)
//   before_agent_start       → user_prompt_submit (per-turn slice) + journal nudge
//   session_before_compact   → pre_compact      (force a full re-ground next turn)
//   session_shutdown         → session_end      (continuity marker)
//
//   * pi's session_start cannot inject into the model, so we stash that context
//     and inject it on the first before_agent_start (which CAN inject).
//
// Requirements: the `sf` binary on PATH (or set SPECFORGE_SF_BIN). Fail-open: any
// adapter error → allow / no injection, never block the user (mirrors the engine).
//
// Install: register this file in pi's settings (~/.pi/agent/settings.json global,
// or .pi/settings.json per-project):
//   { "extensions": ["/abs/path/to/specforge/hooks/pi/specforge.js"] }

import { execFileSync } from "node:child_process";

const SF_BIN = process.env.SPECFORGE_SF_BIN || "sf";

// callHook ejecuta `sf hook --harness=generic` con el payload normalizado por
// stdin y devuelve la decisión parseada. Ante CUALQUIER error → {} (fail open):
// el adapter nunca debe romper el editing por un fallo propio.
function callHook(event, cwd, extra = {}) {
  try {
    const payload = JSON.stringify({ event, project_dir: cwd, ...extra });
    const out = execFileSync(SF_BIN, ["hook", "--harness=generic"], {
      input: payload,
      encoding: "utf8",
      timeout: 5000,
    });
    return out.trim() ? JSON.parse(out) : {};
  } catch {
    return {}; // sf ausente, timeout, JSON inválido… → no interferir
  }
}

export default function (pi) {
  // pi corre esta factory UNA vez por sesión, así que un id generado acá es un
  // identificador de sesión estable — lo necesita el motor para el state machine
  // de inyección del slice (breadcrumb vs slice completo, force-full post-compact).
  const sessionId = `pi-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

  // session_start en pi no puede inyectar al modelo; guardamos su contexto acá y
  // lo soltamos en el primer before_agent_start (que sí inyecta).
  let pendingSessionContext = null;

  // ── tool_call → pre_tool_use (el hard-gate) ──
  // Solo write/edit tienen un path de artefacto gateable; el resto pasa derecho.
  pi.on("tool_call", (event, ctx) => {
    if (event.toolName !== "write" && event.toolName !== "edit") return;
    const filePath = event.input && event.input.path;
    if (!filePath) return;
    const res = callHook("pre_tool_use", ctx.cwd, {
      tool: event.toolName,
      file_path: filePath,
    });
    if (res.decision === "deny") {
      return { block: true, reason: res.reason || "blocked by a SpecForge gate" };
    }
    // allow → no devolvemos nada (dejamos seguir el tool call)
  });

  // ── session_start → guardamos el contexto para inyectarlo el próximo turno ──
  pi.on("session_start", (_event, ctx) => {
    const res = callHook("session_start", ctx.cwd);
    if (res.context) pendingSessionContext = res.context;
  });

  // ── before_agent_start → user_prompt_submit (inyección por turno) ──
  // Componemos un solo mensaje con: (1) el contexto de sesión pendiente, una vez;
  // (2) el slice del paso actual (breadcrumb o slice completo, lo decide el motor);
  // (3) el nudge del journal si una feature quedó archivada sin lecciones.
  pi.on("before_agent_start", (_event, ctx) => {
    const parts = [];

    if (pendingSessionContext) {
      parts.push(pendingSessionContext);
      pendingSessionContext = null; // inyectar una sola vez
    }

    const slice = callHook("user_prompt_submit", ctx.cwd, { session_id: sessionId });
    if (slice.context) parts.push(slice.context);

    // Conciliador de memoria: el motor decide (nudge-once por feature, vía .state).
    const nudge = callHook("stop", ctx.cwd);
    if (nudge.decision === "block" && nudge.reason) parts.push(nudge.reason);

    if (parts.length === 0) return;
    return {
      message: {
        customType: "specforge",
        content: parts.join("\n\n"),
        display: false, // contexto para el modelo, no ruido en la UI
      },
    };
  });

  // ── session_before_compact → pre_compact (re-grounding) ──
  // Marca la sesión para que el próximo before_agent_start inyecte el slice
  // COMPLETO (tras compactar, los slices previos del transcript se resumieron).
  // No cancelamos la compactación.
  pi.on("session_before_compact", (_event, ctx) => {
    callHook("pre_compact", ctx.cwd, { session_id: sessionId });
  });

  // ── session_shutdown → session_end (marca de continuidad) ──
  pi.on("session_shutdown", (_event, ctx) => {
    callHook("session_end", ctx.cwd);
  });
}
