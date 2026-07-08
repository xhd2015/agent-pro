#!/usr/bin/env bash
# Verification plan for message-card UX goal (plan.md steps 1-5).
# Usage: SCRATCH=/path AGENT_RUN_TOKEN=<token> ./script/debug/verify-message-card-ux.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRATCH="${SCRATCH:?SCRATCH must be set}"
mkdir -p "$SCRATCH"

log() { echo "[verify] $*" | tee -a "$SCRATCH/verify.log"; }

: >"$SCRATCH/verify.log"

log "step 1: install"
(cd "$REPO_ROOT" && go run ./script/agent-run/install) 2>&1 | tee "$SCRATCH/install.log"
grep -q 'agent-run installed' "$SCRATCH/install.log"

log "step 2: start agent-run web on :8192"
if lsof -ti:8192 >/dev/null 2>&1; then
  lsof -ti:8192 | xargs kill -9 2>/dev/null || true
  sleep 1
fi
if [[ -z "${AGENT_RUN_TOKEN:-}" ]]; then
  AGENT_RUN_TOKEN="$(openssl rand -hex 16)"
  export AGENT_RUN_TOKEN
  log "generated AGENT_RUN_TOKEN=$AGENT_RUN_TOKEN"
fi
nohup agent-run web --token "$AGENT_RUN_TOKEN" >"$SCRATCH/web-start.log" 2>&1 &
for _ in $(seq 1 40); do
  if curl -sf "http://127.0.0.1:8192/api/agent-run/health" \
    -H "Authorization: Bearer $AGENT_RUN_TOKEN" >"$SCRATCH/health.json"; then
    break
  fi
  sleep 0.5
done
test -s "$SCRATCH/health.json"

log "step 3: playwright probe twice (committed scripts only)"
export SCRATCH AGENT_RUN_TOKEN
"$REPO_ROOT/script/debug/run-message-card-probe-twice.sh" 2>&1 | tee "$SCRATCH/probe-twice.log"
test -s "$SCRATCH/message-card-report.json"
python3 - <<'PY'
import json, os, sys
p = os.path.join(os.environ["SCRATCH"], "message-card-report.json")
with open(p) as f:
    r = json.load(f)
assert r.get("pass") is True, r
assert len(r.get("runs", [])) == 2, r
for run in r["runs"]:
    assert not run.get("issues"), run
print("probe report OK:", p)
PY

log "step 4: go test doctest wrappers (web-layout + enhance-chat)"
(
  cd "$REPO_ROOT/frontend-agent-run" && bun run test
) 2>&1 | tee "$SCRATCH/vitest.log"

(
  cd "$REPO_ROOT"
  go test -count=1 -v ./cmd/agent-run/tests/web-layout/... ./cmd/agent-run/tests/enhance-chat/...
) 2>&1 | tee "$SCRATCH/doctest.log"

grep -q 'PASS (27/27)' "$SCRATCH/doctest.log" || {
  log "web-layout suite did not pass — see doctest.log"
  exit 1
}
cp "$SCRATCH/doctest.log" "$SCRATCH/doctest-web-layout.log"

log "step 4c: critical leaves (multi-tool + poll + grok message cards)"
(
  cd "$REPO_ROOT"
  doctest test --rm -count=1 -v ./cmd/agent-run/tests/web-layout/mobile-progress-multi-tool-ordering --label chromium
  doctest test --rm -count=1 -v ./cmd/agent-run/tests/web-layout/mobile-grok-tty-message-cards --label chromium
  doctest test --rm -count=1 -v ./cmd/agent-run/tests/web-layout/mobile-no-session-detail-poll-while-running --label 'chromium && slow'
) 2>&1 | tee "$SCRATCH/doctest-critical.log"
grep -c 'PASS (1/1)' "$SCRATCH/doctest-critical.log" | grep -q '^3$'

log "step 5: ux-changes evidence"
{
  echo "UX source files (message-card goal):"
  echo "  frontend-agent-run/src/App.tsx"
  echo "  frontend-agent-run/src/App.css"
  echo "  frontend-agent-run/src/progressTimeline.ts"
  echo "  frontend-agent-run/src/progressTimeline.test.ts"
  echo "  script/debug/message-card-ux-probe.js"
  echo "  script/debug/run-message-card-probe-twice.sh"
  echo "  script/debug/verify-message-card-ux.sh"
  echo "Live session URL: http://127.0.0.1:8192/sessions/grok-tty/web_a1e886dbcebb3e2b"
  echo "Probe report: $SCRATCH/message-card-report.json ($(wc -c <"$SCRATCH/message-card-report.json") bytes)"
} | tee "$SCRATCH/ux-changes.txt"

log "verification complete"