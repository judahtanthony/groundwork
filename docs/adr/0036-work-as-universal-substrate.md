# ADR 0036: Work As A Universal Substrate; Software As The Seed Domain

Status: Accepted

## Context

Groundwork's current docs frame the system as coordination for *software* work
(ADR 0006, `vision.md`), with a Symphony-style goal in which humans manage and agents
increasingly execute. That framing is correct for the seed domain but understates the
intended generality. The long-term goal is broader: administer collaboration on *work
of any kind*, and progressively migrate not just execution but **judgment and authority**
from humans into AI agents. Several existing decisions already lean this way — the uniform
work node with advisory `kind` (ADR 0009), org-defined `work_type`, autonomy modeled as
loosenable policy gates (ADR 0006/0011), and canon-as-memory (ADR 0013) — but no ADR
states the destination, so conservative v1 defaults risk being read as the ceiling.

## Decision

Ratify the generalized vision as the north star the architecture serves.

- **Work is a scale-invariant substrate.** A work node carries requirements (including
  resources), procedures, a definition of done, and **value**. The same object models
  "update a doc so future planners understand a decision" and arbitrarily large product or
  organizational work. Software development is the **seed domain**, not the essence; the
  status model and core abstractions must not encode any one domain's process (extends the
  `work_type`-not-status rule, `conventions.md`).

- **The system administers human+AI collaboration and migrates capability *and*
  authority into agents over time**, balancing velocity (agents) against quality (humans).
  Success is humans continuously rising in altitude as procedures, values, and principles
  are encoded into the system — until agents can operate autonomously.

- **Three judgment layers are targets for automation**, mapping to today's human roles:
  1. **Prioritization** — value/judgment ("what is worth doing"; today PM/PO).
  2. **Direction** — translating value into solutions and execution (today design + eng).
  3. **Improvement** — recursive enhancement of the system that identifies and administers
     work itself.

- **Core principle: human authority is a configurable, retractable position on a trust
  gradient, never hardcoded structure.** Human-in-the-loop is a transitional state
  governed by complexity and risk. The system must eventually be able to improve the
  gradient itself (Layer 3). What stays fixed across all autonomy levels is named
  separately (ADR 0037); what loosens is named there too.

This ADR is directional, not implementational. It changes no code, policy, or schema.

## Consequences

ADR 0037 operationalizes this by separating transitional defaults from true invariants;
ADR 0038 begins removing the structural human wirings (reversibility floor, elevation
carve-out) so authority can move along the gradient. The **value/prioritization layer
(Layer 1) has no substrate today** (scheduling is FIFO-by-id; `priority`/`risk_score` are
inert display fields); this ADR names that gap as acknowledged future direction and
**defers** a dedicated value/prioritization ADR to a later phase. `vision.md` is updated
to point here for the long-term goal.
