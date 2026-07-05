# Web UI Information Architecture (v1)

Status: Draft for build. Companion to [ADR 0042](../adr/0042-web-ui-architecture.md)
§"Information Architecture (v1)". This note is the source of truth for the Claude Design
handoff. Scope: the **human** web UI. It excludes the future admin/coordinator-agent
(realtime, streaming, tool-using) client, which is deferred with realtime.

## 1. Purpose and scope

The web UI is Groundwork's primary human interface at full parity with the CLI
([ADR 0041](../adr/0041-human-cli-operating-model.md)). This note fixes the information
architecture — how a human navigates an arbitrary work-node DAG, reviews plans and
implementations at any level of abstraction, and acts through the same gates the CLI uses —
so the visual design can be produced against a settled model. Realtime is deferred
(request/response + light polling now; the API contract stays WebSocket-capable).

## 2. Model recap

- **Uniform work-node.** One node type. Nodes form a tree by parent, plus a **dependency DAG
  overlay** (a node may `depend_on` non-parent nodes). Triage sets `node_type`: `composite`
  (has children) or `leaf` (one verifiable change).
- **Roots vs. envelope-covered subtrees.** **Roots** are parentless nodes, typically
  human-created. Below a node, an approved **envelope** authorizes agents to plan and execute;
  agent-created branch/leaf nodes exist because of that envelope.
- **Lifecycle.** `backlog → todo → in_progress → (blocked) / review → approved → landing →
  done`; plus `cancelled` (and `archived`, once the prune/archive work lands — T-1096).
- **Eligibility.** A node is eligible when `todo` **and** all dependencies are satisfied. The
  eligible set is ordered by **value**: `priority` down the ancestor path, then DFS/FIFO
  ([ADR 0039](../adr/0039-value-and-prioritization.md)). Priority is a **soft** order;
  dependencies are a **hard** gate.
- **Envelopes and the planning budget.** An envelope carries approved actions, allowed roles,
  file scope, risk ceiling, and a **planning budget** (max depth / max children / allowed
  work-types) ([ADR 0044](../adr/0044-hierarchical-planning-and-approval-envelopes.md),
  [0054](../adr/0054-approval-envelopes-v1.md)). A claim/action crossing the envelope raises a
  human **exception** ([ADR 0056](../adr/0056-envelope-aware-claim-authorization.md)).
- **Execution and landing.** The scheduler dispatches an eligible leaf to an AI runtime that
  runs in an isolated git worktree (`gw/run/<id>`), captures changed files + diff, checkpoints,
  writes a completion summary, and moves the node to `review`. `land_to_parent` squashes the
  run branch onto the root integration branch (`gw/root/<id>`); `land_to_main` merges that
  branch and is human-gated ([ADR 0058](../adr/0058-integration-targets-and-landing-levels-v1.md),
  [0059](../adr/0059-worktree-per-run-topology-and-land-from-worktree.md)).
- **Durable records.** Each node has a context brief (`gw ticket context`): ancestor spine,
  parent contract, acceptance, SOP, dependencies, decision/input/handoff records, validation
  state, run evidence, completion/handoff summaries. Review at the parent/root boundary uses
  bulk review bundles ([ADR 0057](../adr/0057-bulk-review-bundles-v1.md)).

## 3. The two surfaces

A single rail of roots does not scale to hundreds of initiatives, so the UI is two surfaces
with a clean handoff: **Roots board (pick an initiative) → node view (work it) → back**.

### Roots board — the landing index

Not a KPI dashboard; a root-scoped, attention-first index (the "board for roots"):

- **Attention-first default:** needs-me + active work surfaced first; **archived roots hidden
  by default** (T-1096) with a "show archived" toggle. At hundreds of roots most are settled,
  so the working set stays small; a count ("6 of 214") keeps the hidden pile honest.
- **Search + global ⌘K command palette:** jump to any **root or node** by id/title from
  anywhere. The palette searches **globally across all roots by default** (cross-tree
  dependency jumps and the many-roots reality make global natural); scoping to a subtree is a
  filter, not the default.
- **Filters × sort × lanes:** filters (needs-me / mine / active / status / owner / label),
  sort (recent activity / attention / priority), optional grouping into lanes.
- **Dense, virtualized/paginated bordered rows:** each row shows root id + title, rollup
  status, progress (settled / total), owner glyph, and right-aligned attention badges
  (approvals · runs · exceptions, the last pinned/red when present).
- KPIs ride as a thin header strip. "New root" creates from here.

### Node view — the drill/work surface

Entered from a root. **No root rail** — only a compact **root switcher** and a **"Roots
board" home button**. It is the up · here · down spine (§7 wireframe W2). Settled/archived
**children** hide by default behind a per-column "show settled" toggle, matching the board.

### Deep-linking

Canonical URLs restore full state: `#/root/:id`, `#/node/:id` (root selection + spine +
children + detail), `#/run/:id`, `#/approval/:id`, `#/node/:id/review`. Deep-linking makes
the UI shareable and lets a notification land the human exactly on the node — the parity
anchor for the CLI's "operate on an id."

## 4. Node-agnostic review model

The detail + action surface operates on **whatever node is focused, at any level of the
DAG** — not just leaves. The detail always follows focus; drilling a child re-points it onto
that child. Every action is expressed against the focused node regardless of depth.

This is the mechanism behind Groundwork's telos of **moving the human up the decision
framework**: early on a human may review down to leaf implementations; as trust accrues the
primary review object shifts upward to plans and designs — approve an envelope on a composite,
approve a decomposition proposal, request a replan. **Progressive elevation is a policy dial,
not a UI rewrite:** via the **envelope planning budget**, a human approves scope once at a
composite/root, agents decompose below without a per-layer gate, and anything past the budget
surfaces as an exception. The same node view, detail panel, and action rail serve a human who
reviews every leaf and one who reviews only the top-level envelope — the difference is the
envelope, not the screen. Cascading decomposition inside a budget is surfaced as a **digest**
("new subtree ready for review") — delivered as approvals-inbox entries plus a Roots-board
attention badge — consistent with realtime being deferred.

## 5. Dependency overlay and priority

Parent→child is the axis you **walk** (the spine); `depends_on` is the axis you **annotate**.
Dependencies are never edges in the children column and never a whole-graph canvas. Three
escalating affordances:

- **Badge** on every row/detail: `blocked-by N` / `blocks N`.
- **Peek** popover: clicking a badge lists the specific `depends_on` nodes with status
  ("waiting on T-0502 · in_progress"), each a deep-link — the per-node `gw ticket list
  --blocked` answer.
- **Cross-tree jump:** following a dependency that points outside the current root's subtree
  switches the active root (with a back affordance). This is the one place the node view
  legitimately leaves its subtree.

**Priority (soft) vs. dependencies (hard).** Dragging children to reorder sets each node's
`priority` — the soft value order the scheduler consumes. Dependencies remain a hard
eligibility gate: a child dragged to the top still won't dispatch until its deps are
satisfied, shown as an **amber "waiting" lock** distinct from the red `blocked` status.

## 6. Provenance and authority

The human/agent boundary in a subtree is drawn where an **approved envelope begins to
authorize agent work** ("below here, an approved envelope lets agents plan and execute"), with
**creator-actor badges** (human vs. AI planner/coding/reviewer) as the per-row visual signal.
Provenance and authority are shown together: an agent-created subtree exists *because* a human
approved an envelope on an ancestor, and the detail panel names that envelope ("governed by
env-E-0006, approved by you").

## 7. Screen / view inventory

Each maps to contract endpoints so parity is checkable.

| View | Surface | Purpose | Primary actions | Backed by |
|---|---|---|---|---|
| **Roots board** | Landing index | Scalable, attention-first root index. | New root; open root; search / ⌘K; filter × sort × lanes; toggle archived | `/tickets` (roots), `/state` |
| **Node view** (up·here·down) | Drill/work | Walk a root's DAG; focus any node; read the brief; act. | Drill up/down; drag-to-prioritize; per-child breakdown; focused-node actions; cross-tree dep jump; root switcher; toggle settled children | `/tickets/:id{,/children,/context,/dependencies}` |
| **Focused-node detail + action rail** | In node view | The `gw ticket context` brief + actions valid at the node's level. | Claim; transition; request breakdown (this node); propose/approve envelope at this level; land (auto-routed); request replan; approve; link dep; reparent; compose decision/input request | `/tickets/:id/context`, `/transition`, `/decompose`, `/envelope`, `/escalate`, `/land*`, `/decisions`, `/dependencies` |
| **Approvals inbox** | Cross-root lens | One queue for all human gates; exceptions pinned; grouped under parent envelope. | Approve / reject / clarify; approve-and-suggest-rule | `/approvals{,/:id/...}` |
| **Decompose / plan review** | From inbox / digest | Review a decomposition proposal: children + parent contract + dep edges + envelope. | Approve (children `backlog→todo` as deps allow); request changes / clarify; reject | `/tickets/:id/decompose`, `/approvals/:id/*` |
| **Envelope editor** | From node detail | Propose/inspect an envelope at the focused composite/root. | Propose (opens `approve_envelope`); inspect active envelope | `/tickets/:id/envelope` |
| **Review bundle** | From root/composite | Feature-level evidence review before `land_to_main`. | Land to main (human-gated); hold; send child to rework | `/tickets/:id/land/preview`, `/land`, bundle (ADR 0057) |
| **Run detail** | From a node/run | One run's mode, plan-or-diff, validations, events, checkpoints, actor snapshot, worktree branch. | Pause / resume / cancel; open workspace; copy resume; export transcript | `/runs/:id{,/events}`, pause/resume/cancel |
| **Run list** | Cross-root lens | All runs; find/triage active + past. | Open a run; filter | `/runs` |
| **Policies** | Config | Trust rules, autonomy ladder, validation templates, SOP maturity, suggestion queue. | Edit rule; elevate autonomy (explicit human act); promote/dismiss suggestion | `/policies{,/suggestions,...}` |
| **Actors** | Config | Registry from `actors.yaml`; roles/capabilities; runtime/model/sandbox. | View/edit; validate | `/actors{,/:id,/validate}` |
| **Settings** | Config | Paths, engine + sandbox, concurrency/lease, server bind/port, AGENTS.md sync, `gw doctor`. | Edit; run doctor | server config; `doctor` |

The old flat map's **Board** (status columns) and **Tickets tree** are subsumed: the Roots
board is the scalable index and the node-view spine *is* the tree. A cross-root status board
survives only as an optional lane view on the Roots board.

## 8. Interaction flows

- **F1 — Land, pick a root, enter the node view.** Board opens attention-first (needs-me +
  active), archived hidden. Search or ⌘K to jump. Selecting a root enters the node view with
  that root centered; the home button and root switcher stay in reach.
- **F2 — Create a root.** "New root" → minimal form (title, kind, priority, optional
  acceptance) → `POST /tickets` (`parent:null`) → node view opens with an empty children
  column prompting "Request breakdown" (and optionally "Propose envelope at this level").
- **F3 — Node-agnostic focus + act.** Detail follows focus. On a leaf: claim / transition /
  land. On a composite/root: propose or approve an envelope, review or request a decomposition,
  request a replan. Drilling a child slides it to center and re-points the detail.
- **F4 — Per-child breakdown.** A child row launches decomposition on **that node alone**
  (`POST /tickets/:id/decompose` → planning run → proposal) without expanding siblings; the
  node → `review`; a run indicator shows; on completion the human is routed to plan review.
- **F5 — Request breakdown → approve → cascading decomposition (digest).** Approving a proposal
  moves children `backlog→todo` as deps allow. If the ancestor envelope authorizes
  `decompose_children` within its planning budget, approved children may be decomposed further
  by agents **with no per-layer approval**; the human receives a **digest** (inbox entry + board
  attention badge), not live growth. Expansion past the budget raises an **exception**.
- **F6 — Drag-to-prioritize children.** Dragging sets each node's `priority` (soft order, ADR
  0039). Dependencies remain a hard gate: a child dragged to the top still won't dispatch until
  its deps are satisfied ("waiting" lock, distinct from `blocked`).
- **F7 — Claim / transition.** Claim verifies eligibility server-side and refuses
  blocked/ineligible nodes with the reason inline. Human claims need no envelope; AI dispatch
  needs trust policy **and** an active envelope (ADR 0056) — the detail says which is missing.
- **F8 — land_to_parent → review bundle → human land_to_main.** Land reads
  `GET /tickets/:id/land/route` and auto-routes: a run-backed child → land to parent (squash
  `gw/run/<id>` onto `gw/root/<id>`); a root → land to main. Child landings accrue on the root
  branch; the review bundle assembles the evidence; Land to main drives the always-human-gated
  `land_to_main` approval. `{override:true}` is a visibly break-glass, audited path.
- **F9 — Blocked run / handoff.** On block (ADR 0051) the agent writes a durable decision/input
  record, releases the lease, node → `blocked`. The UI renders the **blocker first-class** (the
  durable record, the *why*), never a bare red dot. Answering (via the decision panel) records
  the response; the node becomes eligible; a **new** run resumes from a durable resume packet.
- **F10 — Request replan.** On **every** node's action rail: opens an optional feedback note and
  routes an upward revision (`gw ticket escalate`), human-gated in v1.
- **F11 — Compose a decision/input request.** In the node's decision panel, a first-class
  compose flow creates a durable `decision_requested` / `input_requested` record
  (`POST /tickets/:id/decisions`), mapping `gw ticket request`. `export`/`import`/`sync` stay
  CLI-only for v1.
- **F12 — Act on an approval / exception.** The inbox groups by kind under parent envelope.
  Normal items open their context (proposal or diff) with approve / reject / clarify.
  **Exceptions are pinned and elevated**, name which boundary was crossed and the envelope, and
  offer approve-once / reject / widen-the-envelope — keeping the human at boundary crossings only.

## 9. State coverage

All six states specified per list/panel/action:

- **Empty.** Board empty → "Create your first root." Empty children → "Request breakdown."
  Empty inbox → "You're clear." Filtered to archived-only → "No active roots; show archived."
- **Loading.** Request/response + light polling. Board rows show skeletons; node-view panes show
  spinner-with-last-known-value; never blank-flash populated content.
- **Blocked / ineligible.** Two distinct visuals: `blocked` status (red, must name its durable
  blocker) vs. eligible-but-waiting-on-deps (amber "waiting" lock listing unsatisfied deps).
  Ineligible actions render disabled **with the reason inline**, never hidden.
- **Error.** The contract `{error:{code,message}}` envelope renders inline and actionable
  (e.g. `not_a_repo` on a land preview → offer settings). Never a raw stack.
- **Pending-approval.** Badge on the node + count in the header strip + inbox entry — the same
  gate visible in all three, deep-linkable.
- **Active-run.** Indicator on the node, in the children column, in the board row's attention
  badge; planning vs. implementation runs badged differently; opens run detail.

## 10. CLI → UI parity matrix

| CLI group | UI home | Notes |
|---|---|---|
| `status` | Header strip (both surfaces) | Eligible / blocked / pending counts |
| `board` | Roots board (optional status lanes) | Cross-root status is a lane view, not a top page |
| `next` / `--claim` | Node-view "next up" (top of value-ordered children) | `--claim` = one-click |
| `ticket create` | New-root (board) / new-child (node view) | — |
| `ticket list --status/--ready/--blocked` | Board filters + node-view dep peek | eligibility filters |
| `ticket show` / `context` | Focused-node detail | Core |
| `ticket tree` | The node-view spine **is** the tree | — |
| `ticket triage` | Triage banner on a claimed node | leaf/composite |
| `ticket decompose` | Request breakdown (per-node) → plan review | run + approval, not an edit |
| `ticket transition` | Transition action | valid targets only |
| `ticket claim` / `assign` | Claim / assign | server-side eligibility |
| `ticket link --depends-on` | Link-dependency + dep peek | cycle rejection inline |
| `ticket edit --parent` | Reparent | cycle prevention |
| `ticket land [--preview/--to-parent/--all/--override]` | Land (auto-routed) + review bundle | override = break-glass |
| `ticket escalate` | Request replan on every node's rail | — |
| `ticket decisions` / `request` | Decision panel + compose flow | `request` first-class |
| `ticket export` / `import` | — | **CLI-only for v1** |
| `approval list/show/approve/reject/clarify` | Unified inbox (exceptions pinned) | full coverage |
| `run list/show/pause/resume/cancel` | Run list + run detail | human controls only |
| `run once` / `run next` | — | **CLI/scheduler-only**; UI shows effects |
| `validation list/run` | Validation section in detail | `/validations` |
| `envelope propose` | Envelope editor → `approve_envelope` | planning budget = elevation dial |
| `actor list/show/validate` | Actors surface | — |
| `server` | Settings (bind/port/status) | lifecycle via CLI |
| `doctor` | Settings → health | — |
| `export` / `sync` (top-level) | — | **CLI-only for v1** |

No open parity gaps remain for v1: the two exclusions (`export`/`import`/`sync`, machine
dispatch) are deliberate CLI-only decisions.

## 11. Wireframes (in words)

**W1 — Roots board (landing index).** Full-width. A thin **header strip** of KPIs (active runs
· blocked · pending approvals · in review · landed today) — not the focus. Below: a control
bar — left, a search box + ⌘K palette hint; right, filter chips (needs-me / mine / active /
status / owner / label), a sort control (recent activity / attention / priority), an optional
lanes toggle, and a "show archived" switch (off by default). The body is a dense,
virtualized/paginated list of **bordered rows** (not cards): each shows root id + title, a
rollup status dot, a progress readout (settled / total), an owner glyph, and right-aligned
attention badges (approvals · runs · exceptions, the last pinned/red when present). "New root"
sits top-right. Attention-first default floats needs-me + active rows up.

**W2 — Node view (up · here · down spine).** A compact top bar: "Roots board" home button + a
root switcher (no root rail) + the KPI header strip. The main region is a horizontal
three-zone spine:

- **Left — ancestors (up):** a small stack of consumed-breadcrumb chips (`E-0006 › T-0501 …`),
  each clickable to re-center.
- **Center — the focused node (the pointer):** a prominent card with the node id/title and
  status + eligibility + node-type + kind + envelope/autonomy + active-run + pending-approval
  badges, and beneath it the **detail = context brief** (spine, parent contract, acceptance,
  deps, SOP, durable decisions/handoff, validation, run evidence). Alongside it the **action
  rail** whose contents are node-appropriate: on a leaf — claim / transition / land / request
  replan; on a composite/root — propose/approve envelope at this level / request breakdown /
  review bundle / request replan. A **provenance marker** appears in the spine where envelope
  coverage begins, with creator-actor badges per row.
- **Right — children (down/deeper):** a value-ordered, drag-to-reorder column (drag sets
  `priority`); each child row shows leaf/composite, a status/eligibility chip (amber "waiting"
  lock when deps gate it despite high priority), dependency badges (opening a peek popover), a
  creator-actor badge, and a per-row "Request breakdown" affordance. Settled/archived children
  hide behind a per-column "show settled" toggle. Drilling a child slides it left-to-center;
  the old focus becomes the newest ancestor chip.

Focus + context throughout — the DAG is walked, never rendered as a canvas.

## 12. Deferred / realtime-era

Live tree-growth (vs. the digest), the embedded coordinator/admin-agent chat surface, the
SSE→WebSocket upgrade, and reviewer-agent autonomy are deferred with realtime (a later phase).
The IA above is designed to accommodate them without a rewrite.
