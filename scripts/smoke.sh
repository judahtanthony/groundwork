#!/usr/bin/env bash
# End-to-end smoke test for the gw CLI (Phase 1).
# Builds a CGO-free binary and exercises the store-backed command surface in a
# throwaway repo, asserting the behaviors that the unit tests cannot cover at the
# binary level: lazy DB creation, deterministic export, dependency cycle
# rejection, and the doctor health check. Transactional claim/lease concurrency
# is covered by `go test ./internal/store/sqlite`.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
bin="$(mktemp -d)/gw"
work="$(mktemp -d)"

cleanup() { rm -rf "$work"; }
trap cleanup EXIT

fail() { echo "SMOKE FAIL: $*" >&2; exit 1; }

echo "==> building gw (CGO_ENABLED=0)"
( cd "$repo_root" && CGO_ENABLED=0 go build -o "$bin" ./cmd/gw )

cd "$work"

echo "==> gw init"
"$bin" init >/dev/null
[ -f .groundwork/config.yaml ] || fail "init did not create config.yaml"
[ -e .groundwork/state.sqlite ] && fail "init must not create state.sqlite"

echo "==> create + lazy DB"
"$bin" ticket create --title "Build store" >/dev/null
"$bin" ticket triage T-0001 composite >/dev/null
"$bin" ticket create --title "Schema" --parent T-0001 --status todo >/dev/null
"$bin" ticket create --title "Migrations" --parent T-0001 --status todo >/dev/null
[ -f .groundwork/state.sqlite ] || fail "state.sqlite not created lazily"

echo "==> dependency edge + cycle rejection"
"$bin" ticket link T-0003 --depends-on T-0002 >/dev/null
if "$bin" ticket link T-0002 --depends-on T-0003 >/dev/null 2>&1; then
  fail "cycle was not rejected"
fi

echo "==> deterministic export (export twice, diff)"
"$bin" ticket export T-0003 >/dev/null
cp .groundwork/tickets/T-0003/ticket.md "$work/first.md"
"$bin" ticket export T-0003 >/dev/null
diff -u "$work/first.md" .groundwork/tickets/T-0003/ticket.md || fail "export is not deterministic"

echo "==> context brief"
"$bin" context T-0003 --json | grep -q '"id": "T-0002"' || fail "context missing dependency"

echo "==> eligibility via status + transition"
"$bin" ticket transition T-0002 in_progress >/dev/null
"$bin" ticket transition T-0002 done >/dev/null

echo "==> status + board render"
"$bin" status >/dev/null
"$bin" board >/dev/null

echo "==> doctor reports healthy"
"$bin" doctor >/dev/null || fail "doctor reported unhealthy"

echo "==> doctor fails outside a project"
if ( cd "$(mktemp -d)" && "$bin" doctor >/dev/null 2>&1 ); then
  fail "doctor should fail with no project"
fi

echo "SMOKE OK"
