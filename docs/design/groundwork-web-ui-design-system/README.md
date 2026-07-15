# Groundwork Web UI — design reference

The high-fidelity design for the Groundwork human web UI, produced in Claude Design
from the ratified information architecture ([../web-ui-ia.md](../web-ui-ia.md), ADR 0042
§"Information Architecture (v1)"). This is the visual source of truth the SPA build
consumes; it is a **reference**, not production code.

## What's here

| Path | What it is |
|---|---|
| `tokens/tokens.json` | Design tokens (light + dark). The machine-consumable handoff — color, type, radius, space, motion, elevation. Semantic tokens map 1:1 to the IA: `ok/warn/bad/run/idle` states and `human/ai/judge` provenance colors; `accent` is the rationed oxblood. |
| `tokens/groundwork-ui.css` | Self-contained component stylesheet — every token inlined as `--gw-*`, theme-aware via `[data-theme]`. Only external dependency is a Google Fonts import (Inter + IBM Plex Mono). |
| `screens/board.png` | Hero comp W1 — the Roots board (attention-first index, ⌘K search, archived-with-count, root rows with progress + attention/exception badges). |
| `screens/nodeview.png` | Hero comp W2 — the node view up·here·down spine (ancestors left, focused node centered, children right; envelope-coverage provenance boundary; AI·planner creator badge; per-child breakdown; settled toggle; blocked-by peek). |
| `Groundwork Web UI.dc.html` + `support.js` | The interactive design canvas + its renderer. |

## How it's consumed (build wiring)

- **T-1039** scaffolds the Vite SPA (embedded via `go:embed`) that these tokens style.
- **T-1040** builds the design-system components from `tokens/groundwork-ui.css` +
  `tokens/tokens.json` — the two themes track one source of truth.
- **T-1092** (DAG-oriented navigation) and **T-1093** (full CLI-parity CRUD), plus the
  screen leaves under T-1036, implement against the comps and the IA note.

## Brand substitutions to confirm before build

The aesthetic is "Alpine Modernism" — warm-neutral surfaces, charcoal shell, a rationed
oxblood accent. Two working substitutions carried from the source brand system:

- **Inter** stands in for SF Pro (the canonical UI face) — confirm or supply a licensed face.
- **Lucide** is the working UI icon set — confirm or commission a custom set.

## Note

The underlying personal-brand design system this was built on is intentionally **not**
committed here — it is personal/PII content and this repo has a public remote. The
Groundwork tokens above inline every brand value the SPA needs.
