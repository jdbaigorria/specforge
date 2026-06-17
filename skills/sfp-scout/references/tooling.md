# Scout tooling — the research MCP set

`sfp-scout` only works with real evidence, so it depends on research MCP servers.
SpecForge ships an **opinionated default set** in the plugin's `.mcp.json`. They
start automatically when the plugin is enabled; the key-gated ones need an env
var before they work.

| Server | Purpose | Key needed |
|--------|---------|-----------|
| `context7` | Library / framework docs | none |
| `fetch` | Read a specific URL | none |
| `exa` | Web search (competitors, market signals) | `EXA_API_KEY` |
| `github` | Repo / landscape search | `GITHUB_PERSONAL_ACCESS_TOKEN` |

This set is opinionated, not sacred — swap `exa` for Tavily/Brave, or point
`github` at the hosted server, by editing `.mcp.json`. No secrets live in the
repo: `.mcp.json` references env vars (`${EXA_API_KEY}`), which you set in your
shell or the harness's secret store.

## Setup

1. Enable the plugin (the servers are declared already).
2. Export the keys the gated servers need, e.g.:
   ```sh
   export EXA_API_KEY=...                    # exa.ai
   export GITHUB_PERSONAL_ACCESS_TOKEN=...   # github.com/settings/tokens (public_repo scope is enough)
   ```
3. The keyless servers (`context7`, `fetch`) work with no setup.

## Gate behaviour (Step 0)

- **Web search + GitHub available** → full evidence pass.
- **Only keyless servers** → partial pass: docs + URL reads work, but competitor
  search is weak; say so and lean on `model-prior` flags.
- **None available** → do **not** fabricate research. Either tell the user to set
  the keys above and stop, or run an explicit, consented degraded pass where every
  claim is `model-prior` and the brief is stamped "low-evidence."

The point of gating on tooling is honesty: a discovery brief built from the
model's imagination is worse than none, because it *feels* like research.
