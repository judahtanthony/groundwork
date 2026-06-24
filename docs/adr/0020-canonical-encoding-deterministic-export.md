# ADR 0020: Canonical Value Encoding And Deterministic Export

Status: Accepted
Implemented: Implemented

## Context

The schema (`docs/contracts/sqlite-schema.md`) stores timestamps as TEXT and structured fields as JSON-in-TEXT, and `docs/contracts/ticket-export.md` requires deterministic Markdown exports. Round-tripping and diff-friendliness depend on encoding rules the contracts only show by example.

## Decision

- **Timestamps:** RFC3339 in UTC at fixed second precision (`2006-01-02T15:04:05Z`) everywhere — stored, audited, and exported.
- **JSON columns** (`labels_json`, `acceptance_json`, `contract_json`): canonical JSON — object keys sorted, no insignificant whitespace; author order preserved for inherently ordered lists (acceptance criteria, labels).
- **Export:** fixed front-matter key order; fixed section order (Problem, Acceptance Criteria, and — composite only — Design/Contract and Escalations); LF line endings; a single trailing newline; no export-time-only fields. Re-exporting unchanged state is byte-identical.

## Consequences

All store writers route values through canonical encoders. Export is golden-testable (export, re-export, diff) and reimport is lossless. This ADR governs the determinism acceptance criterion of the export work.
