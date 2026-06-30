// SpecForge enforcement adapter for opencode.
//
// A thin translate pair over `sf hook --harness=generic` (see ../README.md): all
// enforcement logic lives in the `sf` CLI; this plugin only maps opencode's hooks
// to the SpecForge contract and back.
//
// SCOPE — be honest about what opencode supports today:
//   ✓ Hard gate:  tool.execute.before throws to block a gate-skipping write.
//   ✓ Re-ground:  experimental.session.compacting pushes the current slice so the
//                 spec survives compaction.
//   ⚠ Per-turn context injection (the breadcrumb/slice every turn, and the
//     SessionStart resume context) is NOT wired: opencode's only injection points
//     are the experimental `chat.*` hooks, whose Message/Part shapes are unstable
//     and can't be constructed reliably without breaking the request. We do not
//     fabricate them. So on opencode the *hard gate* holds, but the anti-context-
//     dilution injection that Claude/pi get is a known gap (see README).
//
// Requirements: the `sf` binary on PATH (or SPECFORGE_SF_BIN). Fail-open: any
// adapter error → allow, never block on our own failure (mirrors the engine).
//
// Install: drop this file in `.opencode/plugin/` (project) or
// `~/.config/opencode/plugin/` (global); opencode auto-discovers *.js plugins.

import { execFileSync } from "node:child_process";

const SF_BIN = process.env.SPECFORGE_SF_BIN || "sf";

// callHook ejecuta `sf hook --harness=generic` con el payload por stdin y
// devuelve la decisión. Ante cualquier error → {} (fail open).
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
    return {};
  }
}

// El plugin: opencode llama esta función con el contexto (project, directory,
// worktree, client, $) y espera de vuelta el objeto de hooks. `directory` es el
// cwd del proyecto — lo capturamos en el closure.
export const SpecForgePlugin = async ({ directory }) => {
  const cwd = directory;

  return {
    // ── tool.execute.before → pre_tool_use (el hard-gate) ──
    // opencode bloquea TIRANDO: el mensaje del Error llega al agente como razón.
    // write/edit traen un path de artefacto; bash trae un comando que puede
    // escribir estado por la puerta lateral (redirect/tee/sed -i) → Capa 1.
    "tool.execute.before": async (input, output) => {
      const args = output.args || {};
      let payload = null;
      if (input.tool === "write" || input.tool === "edit") {
        // opencode usa filePath (camelCase) en write/edit; toleramos variantes.
        const filePath = args.filePath || args.path || args.file_path;
        if (filePath) payload = { tool: input.tool, file_path: filePath };
      } else if (input.tool === "bash") {
        const command = args.command || args.cmd;
        if (command) payload = { tool: input.tool, command };
      }
      if (!payload) return;
      const res = callHook("pre_tool_use", cwd, payload);
      if (res.decision === "deny") {
        throw new Error(res.reason || "blocked by a SpecForge gate");
      }
      // allow → no throw: el tool sigue
    },

    // ── experimental.session.compacting → pre_compact (re-grounding) ──
    // Empujamos el slice actual a output.context para que, tras compactar, el
    // modelo conserve el contexto del paso. (output.context.push acepta strings.)
    "experimental.session.compacting": async (input, output) => {
      callHook("pre_compact", cwd, { session_id: (input && input.sessionID) || "opencode" });
      const slice = callHook("user_prompt_submit", cwd, {
        session_id: (input && input.sessionID) || "opencode",
      });
      if (slice.context && output && Array.isArray(output.context)) {
        output.context.push(slice.context);
      }
    },
  };
};
