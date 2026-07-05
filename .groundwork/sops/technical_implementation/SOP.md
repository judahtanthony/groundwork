# SOP: Technical implementation work

Operating procedure for `work_type: technical_implementation` nodes — a single
verifiable code change that satisfies a leaf's acceptance criteria. It applies to
any actor (human or AI agent, any runtime); it names activities, never a specific
tool or assistant, so every actor can follow it. Groundwork tracks, gates,
validates, and lands the change; this SOP governs how the work itself is done.

## Scope

Use this SOP when the deliverable is a code change: application/library source,
build wiring, or configuration that ships behavior. A leaf is **one verifiable
change** — the smallest unit that satisfies its acceptance criteria and can be
reviewed and landed on its own. If the node cannot be completed as one such change,
it is mis-triaged: escalate for decomposition rather than growing the diff.

Documentation-only changes follow the `documentation` SOP; test-only changes follow
the `test_implementation` SOP. A code change that needs tests to prove it includes
those tests here.

## Operate within your envelope

- **Stay inside the file scope** the governing envelope allows. Needing to touch a
  file outside that scope is a boundary crossing — stop and escalate; do not widen
  the change to work around it (ADR 0056).
- **Never self-elevate past a gate.** Landing, decomposition, policy changes, and
  irreversible actions are human-gated by default. Surface the block; do not loosen
  policy or approve on a human's behalf (ADR 0037/0038).
- **Work in your assigned worktree only.** An isolated run executes in its own git
  worktree; do not reach outside it.

## Procedure

1. **Orient before touching code.** Read the brief: `gw ticket context <id>` gives
   the ancestor spine, the parent contract, acceptance criteria, this SOP, prior
   decisions/handoffs, and the changed-file scope. Restate to yourself what "done"
   means for this leaf in terms of its acceptance criteria.
2. **Understand the code you are about to change.** Locate the symbols involved and
   what depends on them before editing — read the definitions and their call sites,
   and consider what a change would ripple into. Prefer the project's structural
   search over guessing. Understanding blast radius first prevents a "small fix"
   that breaks a caller.
3. **Match the surrounding code.** Follow the conventions already present in the
   files you touch — naming, error handling, structure, comment density, test style.
   New code should read as if the existing authors wrote it. Do not restructure
   unrelated code or introduce a new pattern/dependency when an established one fits.
4. **Make the smallest correct change.** Implement exactly what the acceptance
   criteria require — no speculative abstraction, no drive-by refactors outside the
   node's scope. Handle the real failure/edge cases the change introduces; do not
   leave partial states, swallowed errors, or TODOs standing in for the work.
5. **Prove it with tests.** Add or adjust tests that exercise the new behavior,
   including its failure paths, not just the happy path. A change that cannot be
   demonstrated by a test that would fail without it is not finished (see the
   `test_implementation` SOP for depth).
6. **Verify by running, not by reading.** Build the project, run the relevant tests,
   and — when the change is observable (a command, an endpoint, a screen) — actually
   run it and confirm the behavior. Never report a change as working on inspection
   alone. If the project defines validation commands for these files, they must pass
   (or you must record why, honestly).
7. **Self-review the diff before handing off.** Read your own change end to end as if
   reviewing someone else's, on two axes:
   - **Correctness** — logic errors, off-by-one, nil/unset handling, concurrency,
     resource leaks, mishandled errors, inputs that break it.
   - **Simplicity & reuse** — dead code, duplication, a simpler shape, or existing
     helpers you should have used. Remove anything the change does not need.
   Fix what you find. This is the cheapest review in the loop.
8. **Stage exactly this node's files.** `git add <paths>` for the leaf's change only.
   The index is the ticket-scoped pathspec: at landing Groundwork commits the staged
   files plus the regenerated ticket export, so unrelated edits must not be staged
   and in-scope edits must not be left unstaged (ADR 0034).
9. **Hand off.** Move the node to review (`gw ticket transition <id> review`, or the
   runtime does this on a produced run) and record a completion summary: what
   changed and why, how you verified it, and any residual risk or follow-up. The
   summary is what the human reviewer and the landing gate rely on (ADR 0047).

## When you cannot finish

Do not guess past a real ambiguity, and do not keep a run alive waiting on a human.
If you are blocked — unclear acceptance, a needed decision above your authority, a
required file outside scope, or a failing dependency — checkpoint your work, write
run evidence and a durable handoff/decision record, move the node to an explainable
`blocked` state with a clear statement, release the lease, and let the scheduler
pick up other work (ADR 0047/0051). A precise blocker is more valuable than a guess.

## Validation and landing

Landing enforces the validation template that matches the changed files (e.g. the
project's build/test commands) before the change is committed. `land_to_parent`
squashes the run branch onto the root integration branch; `land_to_main` (the merge
to the default branch) stays human-gated in v1. "Landed" means "committed" — a green
validation gate plus the human approval where required (ADR 0034/0058).

## Canon and the feedback loop

When a change establishes or contradicts a durable decision, its home is the
canon — an ADR, an architecture or contract doc, or a policy — edited in place, not
left in a node journal (ADR 0013). If the brief lacked something you needed, record
a context-miss (`gw ticket context <id> --miss "…"`) so the brief and this SOP
improve over time (ADR 0035).
