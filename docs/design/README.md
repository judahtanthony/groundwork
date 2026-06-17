# Design

Visual/UI design reference for the Groundwork web app surface.

This directory holds **design reference only** — it is not app code and not a frontend build. The AGENTS.md boundary still applies: do not turn these prototypes into generated frontend assets until web-surface implementation is explicitly started. When that time comes, recreate the visual output in Go server-rendered HTML (per [overview.md](../architecture/overview.md)); do not copy the prototype's React/Babel internals.

## Contents

- [wireframes/](wireframes/) — the Claude Design (claude.ai/design) handoff bundle: HTML/CSS/JSX prototypes of all seven screens plus a component inventory. See [wireframes/ORIGINAL-HANDOFF.md](wireframes/ORIGINAL-HANDOFF.md) for the original handoff instructions. Open `wireframes/index.html` in a browser to view the canvas (it loads React/Babel from a CDN).
- [decomposition-ui-spec.md](decomposition-ui-spec.md) — concrete spec for the work-node / decomposition surfaces the wireframes do **not** yet draw. The wireframes predate the model in ADRs 0009–0011; this spec is the source of truth for closing that gap.

## How the wireframes relate to the durable design

The wireframes are faithful to the v1 state machine, run/approval/validation model, and local-first posture (see the alignment notes folded into [../architecture/dashboard.md](../architecture/dashboard.md)). Three reconciliations are already settled in the docs and override the prototype where they differ:

- Config/policy files are **YAML**, not the TOML the prototype labels show.
- Risk is a **0–100 score mapped onto the four named classes** (`low`/`medium`/`high`/`critical`); gates key off the class.
- Reviewer-agent modes are **Phase 2**; v1 exposes only auto and require-human.

The largest gap — the entire work-node hierarchy, dependencies, decomposition proposals, escalation, SOPs, and autonomy levels — is specified in [decomposition-ui-spec.md](decomposition-ui-spec.md).
