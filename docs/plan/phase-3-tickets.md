# Phase 3 (M3 — Self-Host Low-Risk Work) Ticket Breakdown

> **Status: implemented.** M3 (T-1001, T-1002, T-1004, T-1006, T-1007, T-1008) is
> built, tested (unit + `-race` + `scripts/smoke.sh`), and dogfooded: the work tree is
> imported and a real docs ticket (T-1002) was driven through Groundwork's own gates,
> with gw making the landing commit. T-1003 (first Codex-assisted ticket) remains
> deferred to Phase 4. Dependency-ordered breakdown for **M3: Self-Hosting
> Preparation** (see `docs/product/roadmap.md`, `docs/plan/milestones.md`,
> `docs/architecture/dogfooding.md`). New ADRs: 0032–0035.

Ticket ids continue the `work-tree.yaml` convention; new tickets fill gaps within
epic **E-0011**'s `10xx` range after the planning-defined T-1001/1002/1003 and are
marked **(new)**. Each ticket cross-references its `work-tree.yaml` epic/ticket and
the governing ADR(s). All tickets are children of **E-0011 — Dogfood Groundwork on
Groundwork**.

## Scope boundary (decided)

**In M3:** import the bootstrap work tree (E-0011/T-1001), then take real low-risk
**docs** tickets through Groundwork as the system of record + gates (E-0011/T-1002),
with human landing approval. Execution is **human-performed** via manual status
transitions ([ADR 0033](../adr/0033-human-execution-via-manual-transitions.md)); gw
owns the landing commit via a minimal `internal/git`
([ADR 0034](../adr/0034-minimal-git-landing.md)); the dogfooding feedback loop gets
an observable signal ([ADR 0035](../adr/0035-context-miss-capture.md)). M3 stays on
the records-only runtime stub and the localhost, single-user server.

**Deferred:**
- **T-1003 — first Codex-assisted ticket → Phase 4 / M4.** Its acceptance ("Codex
  run is tracked", "human landing gate exercised") requires the real Codex runtime
  (isolated worktrees, run-event streaming, checkpoint/squash/resume git). This is
  also where human-triggered **run-based** dispatch finally drives a real executor.
  It stays in E-0011, tagged M4. (Phase mapping lives here, not in `work-tree.yaml`,
  matching the Phase 2 precedent.)
- The Codex runtime, isolated per-run worktrees, run-event streaming/transcripts, and
  worktree checkpoint/squash/resume git → Phase 4
  ([ADR 0027](../adr/0027-run-lifecycle-and-checkpoint-records.md)).
- Autonomy elevation → Phase 5. Dashboard HTML/frontend, multi-human roles, auth,
  remote mode → post-v1 (AGENTS.md).

## Wave 0 — Landing-to-git + import prerequisites (no deps; build first)

### T-1004 (new) — Minimal git-landing: gw owns the landing commit
- **Epic:** E-0011. **ADR:** [0034](../adr/0034-minimal-git-landing.md) (composes
  [0028](../adr/0028-gate-evaluation-engine.md) landing gate,
  [0027](../adr/0027-run-lifecycle-and-checkpoint-records.md) boundary).
- **Dep:** none.
- **Acceptance:** new `internal/git` with a single capability — stage a ticket-scoped
  pathspec + the regenerated ticket export and `git commit` on the **current branch**
  with a ticket-referencing message, returning the SHA (recorded on the
  approval/audit trail). `gw ticket land <id>`, after the `land_to_main` approval
  passes, performs this commit and moves the ticket to `done` atomically (gw `done` ⇔
  git commit). Warns/refuses if the working tree carries changes outside the ticket
  pathspec. **No worktrees, branches, checkpoints, squash, or resume.**
- **Gate:** `go test ./internal/git/...` and the store landing tests; integration —
  approving a landing yields the working-tree change + updated export in one commit
  on the branch; `gw ticket show` reports `done`; dirty-tree guard test.

### T-1006 (new) — Importer hardening for whole-tree historical import
- **Epic:** E-0011. **ADR:** [0019](../adr/0019-uniform-ticket-ids.md),
  [0032](../adr/0032-bootstrap-import-via-authored-markdown.md).
- **Dep:** none (parallel with T-1004).
- **Acceptance:** the importer ingests nodes with `status: done`, goal/epic non-`T-`
  parent ids (`G-0001`/`E-00NN`), and cross-subtree `depends_on`; the `T-NNNN`
  allocator reseeds to `max(T-NNNN) + 1`; round-trip (`export` → delete
  `state.sqlite` → `import`) reproduces an identical tree + edges + statuses
  (extends the T-0902 gate to the full tree). If M2 already handles all of this, this
  degrades to a verifying test.
- **Gate:** `go test ./internal/store/...`; round-trip test over the authored
  bootstrap fixture.

## Wave 1 — Import the bootstrap work tree

### T-1001 — Import bootstrap work tree into Groundwork
- **Epic:** E-0011 (`T-1001`). **ADR:**
  [0032](../adr/0032-bootstrap-import-via-authored-markdown.md).
- **Dep:** T-1006.
- **Acceptance:** every `G`/`E`/`T` node in `docs/plan/work-tree.yaml` has a committed
  `.groundwork/tickets/**/ticket.md` export preserving parent/child hierarchy and
  acceptance; Phase 0–2 nodes carry `status: done`, E-0011 + Phase 4–5 nodes carry
  `backlog`/`todo`; **ids preserved (no remap)**; `gw ticket import` rebuilds the
  identical managed tree; `gw ticket tree` shows the full hierarchy. *(Satisfies the
  planning acceptance: "work-tree.yaml becomes managed Groundwork tickets" +
  "imported records preserve hierarchy.")*
- **Gate:** `gw ticket import` then `gw ticket tree --json` matches the authored
  fixture; `bash scripts/smoke.sh` import round-trip passes.

## Wave 2 — Run a low-risk docs ticket end-to-end

### T-1007 (new) — Docs-work SOP + context-brief check
- **Epic:** E-0011. **ADR:** [0013](../adr/0013-canon-as-memory.md),
  [0011](../adr/0011-progressive-planning-autonomy-via-sops.md).
- **Dep:** T-1001.
- **Acceptance:** a docs SOP exists under `.groundwork/sops/documentation/`;
  `gw ticket context <docs-node>` surfaces it alongside the ancestor spine, parent
  contract, and acceptance; the brief stays bounded.
- **Gate:** `go test ./internal/contextbrief/...`; `gw ticket context` on the chosen
  docs node shows the SOP.

### T-1002 — Use Groundwork for a low-risk documentation ticket
- **Epic:** E-0011 (`T-1002`). **ADR:**
  [0032](../adr/0032-bootstrap-import-via-authored-markdown.md),
  [0033](../adr/0033-human-execution-via-manual-transitions.md),
  [0034](../adr/0034-minimal-git-landing.md).
- **Dep:** T-1004, T-1001, T-1007.
- **Acceptance:** a real docs-only node travels `todo → in_progress → review →
  approved → landed`, human-performed: the human edits the doc in the working tree and
  drives `gw ticket transition` (`todo`→`in_progress`→`review`); landing goes through
  the `land_to_main` human gate (`gw ticket land` → `gw approval approve`) with the
  `documentation` validation template (`landing_risk_floor: low`); **gw commits the
  doc change + regenerated export in one commit** (T-1004) and the ticket reaches
  `done`; `runs/`/`approvals/`/`state.sqlite*` stay git-ignored. *(Satisfies the
  planning acceptance: "a docs-only ticket is completed through Groundwork" +
  "runtime logs remain ignored.")*
- **Gate:** captured end-to-end run; `git log -1` shows the single landing commit
  (doc + export); `git status` shows no runtime state staged; `gw ticket show <id>`
  reports `done`.

## Wave 3 — Operationalize the dogfooding feedback loop

### T-1008 (new) — Context-miss capture for the canon loop
- **Epic:** E-0011. **ADR:** [0013](../adr/0013-canon-as-memory.md),
  [0035](../adr/0035-context-miss-capture.md).
- **Dep:** T-1002.
- **Acceptance:** a minimal records-only way to log a context-miss against a node
  (what the worker needed that the brief lacked — e.g.
  `gw ticket context <id> --miss "<note>"` appending to the node journal / an ignored
  log) plus a documented review step that turns recurring misses into a canon/SOP/brief
  edit; at least one real miss from T-1002 captured and resolved into a doc/SOP change.
- **Gate:** `go test` for the capture path; worked example — miss recorded → SOP/brief
  updated → re-running `gw ticket context` shows the gap closed.

## Deferred (tracked in E-0011, not M3)

### T-1003 — Use Groundwork for first Codex-assisted implementation ticket → Phase 4 (M4)
Requires the real Codex runtime (worktrees, run-event streaming, checkpoint/squash/
resume git) and is where human-triggered run-based dispatch first drives a real
executor. See `docs/plan/phase-4-tickets.md` (future).

---

## Build-first & end-to-end verification

1. **First:** T-1004 (git-landing) and T-1006 (importer hardening) in parallel — they
   unblock durable landing and the import.
2. **Import:** author the Markdown exports from `work-tree.yaml`; `gw ticket import`
   → full managed tree (`gw ticket tree`), completed work `done`, hierarchy + deps
   intact (T-1001).
3. **Brief:** pick a genuine low-risk docs node; `gw ticket context <id>` returns a
   usable brief with the docs SOP (T-1007).
4. **Self-host slice (the capstone, T-1002):** `gw ticket transition <id>
   in_progress` → human edits the doc → `gw ticket transition <id> review` →
   `gw ticket land <id>` opens the `land_to_main` human gate → `gw approval approve
   <id>` enforces the `documentation` validation gate → **gw commits the doc change +
   regenerated export in one commit** on the current branch and the ticket reaches
   `done`. Verify `git log -1` shows the single landing commit and `git status` shows
   no `runs/`/`approvals/`/`state.sqlite*` staged.
5. **Feedback loop:** record one context-miss observed during step 4 and resolve it
   into a canon/SOP edit (T-1008).

Wire the slice into a `scripts/smoke.sh` self-host section plus one integration test,
so the self-hosting path is regression-covered before the Codex runtime arrives in
Phase 4.

**Global build gates** (per `docs/reference/agent-quickstart.md`) apply to every code
ticket: `gofmt -l` clean, `CGO_ENABLED=0 go vet ./...`, `go build ./...`,
`go test ./...`, `CGO_ENABLED=1 go test -race ./...` on concurrency paths, and
`bash scripts/smoke.sh`.
