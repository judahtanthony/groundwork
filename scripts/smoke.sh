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
[ -f .groundwork/actors.yaml ] || fail "init did not create actors.yaml"
[ -e .groundwork/state.sqlite ] && fail "init must not create state.sqlite"

echo "==> actor registry"
"$bin" actor validate >/dev/null || fail "actor validate failed"
"$bin" actor list --json | grep -q '"id": "ai.codex.default"' || fail "actor list missing default Codex actor"

echo "==> create + lazy DB"
"$bin" ticket create --title "Build store" >/dev/null
"$bin" ticket triage T-0001 composite >/dev/null
"$bin" ticket create --title "Schema" --parent T-0001 --status todo --work-type technical_implementation >/dev/null
"$bin" ticket create --title "Migrations" --parent T-0001 --status todo --requested-actor ai.codex.default >/dev/null
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

echo "==> ticket export + import round-trip (T-0902)"
for id in T-0001 T-0002 T-0003; do "$bin" ticket export "$id" >/dev/null; done
rm -f .groundwork/state.sqlite .groundwork/state.sqlite-wal .groundwork/state.sqlite-shm
"$bin" ticket import >/dev/null || fail "import failed"
"$bin" ticket show T-0001 --json | grep -q '"node_type": "composite"' || fail "import lost node_type"
"$bin" context T-0003 --json | grep -q '"id": "T-0002"' || fail "import lost dependency edge"

echo "==> status + board render"
"$bin" status >/dev/null
"$bin" board >/dev/null

echo "==> doctor reports healthy"
"$bin" doctor >/dev/null || fail "doctor reported unhealthy"

echo "==> doctor fails outside a project"
if ( cd "$(mktemp -d)" && "$bin" doctor >/dev/null 2>&1 ); then
  fail "doctor should fail with no project"
fi

echo "==> gw server health + state (T-0401)"
# Bind an ephemeral port to avoid clashing with a real coordinator on 4500.
"$bin" server --addr 127.0.0.1:4599 >"$work/server.log" 2>&1 &
server_pid=$!
# Wait for the listener to come up (the server logs its bound address).
for _ in $(seq 1 50); do
  curl -fsS "http://127.0.0.1:4599/healthz" >/dev/null 2>&1 && break
  sleep 0.1
done
api() { curl -fsS "http://127.0.0.1:4599$1"; }
die() { kill "$server_pid" 2>/dev/null; fail "$1"; }
api "/healthz" | grep -q '"status": "ok"' || die "server health not ok"
api "/api/v1/state" | grep -q '"ok": true' || die "server state not ok"
api "/api/v1/state" | grep -q '"counts"' || die "server state missing counts"

echo "==> gw server read API (T-0411)"
api "/api/v1/tickets" | grep -q '"id": "T-0001"' || die "tickets list missing T-0001"
api "/api/v1/tickets/T-0003" | grep -q '"title": "Migrations"' || die "ticket get wrong body"
api "/api/v1/tickets/T-0003/dependencies" | grep -q '"T-0002"' || die "dependencies missing T-0002"
api "/api/v1/tickets/T-0003/context" | grep -q '"id": "T-0002"' || die "context missing dependency"
api "/api/v1/actors" | grep -q '"id": "ai.codex.default"' || die "actors missing default Codex"
# Missing ticket returns a JSON 404, not a 200 or a panic.
code="$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:4599/api/v1/tickets/T-9999")"
[ "$code" = "404" ] || die "missing ticket returned $code, want 404"

echo "==> gw server mutating API (T-0412)"
post() { curl -fsS -X POST -H 'Content-Type: application/json' -d "$2" "http://127.0.0.1:4599$1"; }
status_for() { curl -s -o /dev/null -w '%{http_code}' -X "$1" -d "$3" "http://127.0.0.1:4599$2"; }
new_id="$(post '/api/v1/tickets' '{"title":"Created via API","status":"todo"}' | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -1)"
[ -n "$new_id" ] || die "create via API returned no id"
api "/api/v1/tickets/$new_id" | grep -q '"title": "Created via API"' || die "created ticket not retrievable"
post "/api/v1/tickets/$new_id/transition" '{"status":"in_progress"}' | grep -q '"status": "in_progress"' || die "transition via API failed"
# An illegal transition (in_progress -> todo) returns 409.
code="$(status_for POST "/api/v1/tickets/$new_id/transition" '{"status":"todo"}')"
[ "$code" = "409" ] || die "illegal transition returned $code, want 409"
# Add then remove a dependency through the API.
post "/api/v1/tickets/$new_id/dependencies" '{"depends_on":"T-0001"}' >/dev/null || die "add dependency via API failed"
api "/api/v1/tickets/$new_id/dependencies" | grep -q '"T-0001"' || die "dependency not recorded"
curl -fsS -X DELETE "http://127.0.0.1:4599/api/v1/tickets/$new_id/dependencies/T-0001" >/dev/null || die "remove dependency via API failed"

echo "==> gw coordinator auto-schedules a run (T-0402/0420/0423)"
sched_id="$(post '/api/v1/tickets' '{"title":"Scheduled work","status":"todo","work_type":"technical_implementation"}' | sed -n 's/.*"id": "\([^"]*\)".*/\1/p' | head -1)"
[ -n "$sched_id" ] || die "create scheduled ticket failed"
# The scheduler loop should claim it, run the stub, and land it in review.
ok=""
for _ in $(seq 1 50); do
  st="$(api "/api/v1/tickets/$sched_id" | sed -n 's/.*"status": "\([^"]*\)".*/\1/p' | head -1)"
  if [ "$st" = "review" ]; then ok=1; break; fi
  sleep 0.2
done
[ -n "$ok" ] || die "scheduled ticket did not reach review"
api "/api/v1/runs" | grep -q "\"ticket_id\": \"$sched_id\"" || die "no run recorded for scheduled ticket"

echo "==> gw decompose proposal + approve (T-0430)"
parent_id="$(post '/api/v1/tickets' '{"title":"Composite","status":"todo"}' | sed -n 's/.*"id": "\(T-[0-9]*\)".*/\1/p' | head -1)"
[ -n "$parent_id" ] || die "create composite failed"
"$bin" ticket triage "$parent_id" composite >/dev/null || die "triage failed"
appr_id="$(post "/api/v1/tickets/$parent_id/decompose" '{"contract":{"schema":"c/v1"},"children":[{"title":"child a"}]}' | sed -n 's/.*"id": "\(A-[0-9]*\)".*/\1/p' | head -1)"
[ -n "$appr_id" ] || die "decompose proposal failed"
api "/api/v1/tickets/$parent_id" | grep -q '"status": "review"' || die "parent not in review after proposal"
post "/api/v1/approvals/$appr_id/approve" '{"reason":"ok"}' | grep -q '"status": "approved"' || die "approve decompose failed"

echo "==> gw escalate + replan (T-0431)"
esc_id="$(post '/api/v1/tickets' '{"title":"Needs replan","status":"todo"}' | sed -n 's/.*"id": "\(T-[0-9]*\)".*/\1/p' | head -1)"
esc_appr="$(post "/api/v1/tickets/$esc_id/escalate" '{"reason":"changed"}' | sed -n 's/.*"id": "\(A-[0-9]*\)".*/\1/p' | head -1)"
[ -n "$esc_appr" ] || die "escalate failed"
api "/api/v1/tickets/$esc_id" | grep -q '"status": "blocked"' || die "escalated ticket not blocked"
post "/api/v1/approvals/$esc_appr/approve" '{}' | grep -q '"status": "approved"' || die "replan approve failed"
api "/api/v1/tickets/$esc_id" | grep -q '"status": "todo"' || die "replan did not requeue node"

echo "==> gw landing through the gate (T-0602/0603/0702/0703)"
land_id="$(post '/api/v1/tickets' '{"title":"Landable","status":"todo"}' | sed -n 's/.*"id": "\(T-[0-9]*\)".*/\1/p' | head -1)"
[ -n "$land_id" ] || die "create landable failed"
"$bin" ticket transition "$land_id" in_progress >/dev/null || die "transition in_progress"
"$bin" ticket transition "$land_id" review >/dev/null || die "transition review"
api "/api/v1/tickets/$land_id/validations" >/dev/null || die "validations endpoint failed"
# Landing opens a human land_to_main approval (the gate); approving it lands.
land_resp="$(post "/api/v1/tickets/$land_id/land" '{}')"
echo "$land_resp" | grep -q '"landed": false' || die "land should open a pending approval, not land directly"
land_appr="$(echo "$land_resp" | sed -n 's/.*"id": "\(A-[0-9]*\)".*/\1/p' | head -1)"
[ -n "$land_appr" ] || die "land did not open an approval"
post "/api/v1/approvals/$land_appr/approve" '{}' | grep -q '"status": "approved"' || die "land approve failed"
api "/api/v1/tickets/$land_id" | grep -q '"status": "done"' || die "node not landed after approval"

echo "==> gw SSE stream (T-0803)"
sse_out="$(curl -sN --max-time 1 "http://127.0.0.1:4599/api/v1/events" 2>/dev/null || true)"
echo "$sse_out" | grep -q "connected" || die "SSE stream did not connect"

kill "$server_pid" 2>/dev/null
wait "$server_pid" 2>/dev/null || true

echo "SMOKE OK"
