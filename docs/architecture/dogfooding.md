# Dogfooding Groundwork

Using Groundwork to build Groundwork is a good idea, but not immediately.

## Recommendation

Do not use Groundwork for active self-hosted agent work until the coordinator exists. Before that point, use committed docs and the bootstrap work tree as the source of planning truth. (As of M3 the coordinator exists and Groundwork is the planning source of truth — ADR 0040.)

## Dogfooding Phases

1. Bootstrap durable project knowledge.
2. Build CLI and SQLite store from docs.
3. Build coordinator.
4. Import the bootstrap work tree into Groundwork.
5. Use Groundwork for low-risk docs and CLI tickets.
6. Add Codex runtime and use Groundwork for implementation tickets.
7. Enable more autonomy only after validation and trust policies mature.

