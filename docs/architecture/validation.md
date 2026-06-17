# Validation

Validation is first-class and integrated with trust policy.

## Template Model

Validation templates map changed files to required checks. Applicable validation must pass before landing unless a human records an explicit override.

Example template categories:

- Documentation: Markdown and internal guidance files.
- Go: `**/*.go`.
- Web: package and frontend source files.

## Example Rules

Documentation-only changes:

- May have no required command if no formatter is configured.
- Risk floor may be low.
- May be auto-approved when trust policy permits.

Go changes:

- Require `go test ./...`.

Web changes:

- Require `npm test`.
- Require `npm run typecheck` when configured.

## Landing Gate

Landing to `main` requires applicable validation to pass and the landing approval gate to permit the action.

