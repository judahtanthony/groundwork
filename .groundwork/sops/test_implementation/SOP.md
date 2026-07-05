# SOP: Test implementation work

Operating procedure for `work_type: test_implementation` nodes — a verifiable change
whose deliverable is automated tests. It applies to any actor (human or AI agent,
any runtime) and names activities, not tools. Groundwork tracks, gates, validates,
and lands the change; this SOP governs how good tests are written.

## Scope

Use this SOP when the node's deliverable is test code: unit, integration, or
end-to-end tests for existing or newly added behavior. When tests are part of
shipping a feature, write them under the `technical_implementation` SOP with the
code; use this SOP for nodes whose whole purpose is test coverage (backfilling a
gap, an e2e harness, a regression suite). A leaf is **one verifiable change**.

The envelope, gate, worktree, staging, and landing rules in the
`technical_implementation` SOP apply here unchanged — stay in file scope, never
self-elevate, work in your worktree, stage only this node's files.

## What makes a good test (the bar to meet)

1. **It can fail.** A test that passes against broken code is worse than none. Before
   trusting a test, confirm it fails without the behavior it covers (write it first,
   or temporarily break the code) and then passes with it. Assert on observable
   behavior and outputs, not on incidental implementation detail that will churn.
2. **It covers the failure paths.** The happy path is the easy half. Cover the
   boundaries the change introduces: empty/nil/zero, malformed input, error returns,
   concurrency where relevant, and the specific regression a bug fix addresses.
3. **It is deterministic.** No dependence on wall-clock timing, network, ordering of
   maps/sets, or ambient machine state. Seed randomness, inject clocks, use
   temporary directories. A flaky test erodes the whole gate's credibility.
4. **It matches the project's testing conventions.** Use the framework, helpers,
   fixtures, table-driven style, and naming already in the codebase. New tests should
   read like the existing suite, and live where the suite expects them.
5. **It is fast and focused.** Prefer the smallest scope that proves the behavior
   (unit over integration over e2e when either suffices). One clear reason to fail
   per test; a failure message that points at the cause.

## Procedure

1. **Orient.** `gw ticket context <id>` for the brief, acceptance, this SOP, and the
   changed-file scope. Identify exactly which behaviors this node must cover.
2. **Read the code under test and the existing tests** around it, so new tests fit
   the established patterns and you understand the real edge cases.
3. **Write the tests to the bar above.** Prove each one can fail before it passes.
4. **Run the full relevant suite**, not just the new tests — confirm you did not
   break neighbors and that the new tests pass deterministically (run them more than
   once if timing is involved). The project's validation commands must pass.
5. **Self-review.** Check each test actually asserts the intended behavior (not a
   tautology), the failure paths are covered, and there is no flakiness or dead
   scaffolding. Remove anything unneeded.
6. **Stage this node's files**, hand off to review, and record a completion summary:
   what is now covered, how you confirmed the tests fail-then-pass, and any coverage
   gaps intentionally left (with why).

## When you cannot finish

Follow the `technical_implementation` SOP's blocked-handoff procedure: checkpoint,
write run evidence and a durable handoff record, move to an explainable `blocked`
state, release the lease. If a test reveals a genuine defect in the code under test,
that is a finding worth surfacing (a decision/handoff record or an escalation), not
something to paper over by weakening the test.

## Validation, landing, and canon

Landing enforces the matching validation template and the human gate where required
(ADR 0034/0058), same as any change. If writing tests exposes a missing convention
or a brittle seam worth documenting, record it in the canon and, if the brief lacked
it, log a context-miss so future test nodes start better-informed (ADR 0013/0035).
