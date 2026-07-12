#!/usr/bin/env bash
# Demo: agent-run run --session-id → agent-run sessions list (same store).
#
# Shows:
#   1. CLI run with a stable --session-id creates a session under AGENT_RUN_HOME
#   2. agent-run sessions (and --json) can list that session
#   3. On-disk layout: sessions/<runner>/<session_id>/
#   4. Why runner is part of the key: same session_id under two runners can coexist
#
# Usage (from repo root):
#   ./script/debug/demo-run-then-sessions.sh
#   KEEP_HOME=1 ./script/debug/demo-run-then-sessions.sh   # leave AGENT_RUN_HOME for inspection
#   AGENT_RUN_BIN=/path/to/agent-run ./script/debug/demo-run-then-sessions.sh
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

SESSION_ID="${SESSION_ID:-demo-cli-1}"
RUNNER="${RUNNER:-fake-codex}"
PROMPT="${PROMPT:-say hi from demo-run-then-sessions}"
KEEP_HOME="${KEEP_HOME:-0}"

TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/agent-run-demo-run-sessions.XXXXXX")"
BIN_DIR="${TMP_ROOT}/bin"
HOME_DIR="${TMP_ROOT}/home"
mkdir -p "$BIN_DIR" "$HOME_DIR"

cleanup() {
  if [[ "$KEEP_HOME" == "1" ]]; then
    echo
    echo "KEEP_HOME=1: left AGENT_RUN_HOME at:"
    echo "  $HOME_DIR"
    echo "  (bin dir: $BIN_DIR)"
    return
  fi
  rm -rf "$TMP_ROOT"
}
trap cleanup EXIT

section() {
  echo
  echo "======== $* ========"
}

build_bins() {
  section "build agent-run + fake-codex"
  if [[ -n "${AGENT_RUN_BIN:-}" ]]; then
    cp "$AGENT_RUN_BIN" "$BIN_DIR/agent-run"
    chmod +x "$BIN_DIR/agent-run"
    echo "using AGENT_RUN_BIN=$AGENT_RUN_BIN"
  else
    go build -o "$BIN_DIR/agent-run" ./cmd/agent-run
  fi
  go build -o "$BIN_DIR/fake-codex" ./cmd/fake-codex
  export PATH="$BIN_DIR:$PATH"
  export AGENT_RUN_HOME="$HOME_DIR"
  echo "AGENT_RUN_HOME=$AGENT_RUN_HOME"
  echo "PATH bin=$BIN_DIR"
}

run_with_session_id() {
  section "1) agent-run run --session-id=$SESSION_ID --agent-runner=$RUNNER"
  set -x
  agent-run run \
    --json \
    --agent-runner "$RUNNER" \
    --session-id "$SESSION_ID" \
    "$PROMPT"
  set +x
}

list_sessions() {
  section "2) agent-run sessions (human)"
  agent-run sessions || true

  section "3) agent-run sessions --json"
  agent-run sessions --json

  section "4) on-disk layout under sessions/"
  if [[ -d "$AGENT_RUN_HOME/sessions" ]]; then
    find "$AGENT_RUN_HOME/sessions" -type f | sort | sed "s|^$AGENT_RUN_HOME/||"
    if [[ -f "$AGENT_RUN_HOME/sessions/$RUNNER/$SESSION_ID/meta.json" ]]; then
      echo
      echo "meta.json:"
      python3 -m json.tool "$AGENT_RUN_HOME/sessions/$RUNNER/$SESSION_ID/meta.json" 2>/dev/null \
        || cat "$AGENT_RUN_HOME/sessions/$RUNNER/$SESSION_ID/meta.json"
    fi
  else
    echo "(no sessions/ directory)"
  fi
}

assert_listed() {
  section "5) assert list contains $RUNNER/$SESSION_ID"
  local out
  out="$(agent-run sessions --json)"
  if ! python3 - "$out" "$RUNNER" "$SESSION_ID" <<'PY'
import json, sys
raw, runner, sid = sys.argv[1], sys.argv[2], sys.argv[3]
data = json.loads(raw)
sessions = data.get("sessions") or []
for s in sessions:
    if s.get("runner") == runner and s.get("session_id") == sid:
        print(f"OK: found runner={runner!r} session_id={sid!r} status={s.get('status')!r}")
        sys.exit(0)
print("FAIL: session not listed")
print("sessions:", json.dumps(sessions, indent=2))
sys.exit(1)
PY
  then
    echo "FAIL: expected session not in list" >&2
    exit 1
  fi
}

# Same session_id under a second runner — storage is keyed by (runner, session_id).
demo_runner_namespace() {
  section "6) why runner is part of the key (optional second runner)"
  local other_runner="opencode"
  # Seed only meta (no full opencode run) so the list shows two rows with same session_id.
  local dir="$AGENT_RUN_HOME/sessions/$other_runner/$SESSION_ID"
  mkdir -p "$dir"
  python3 - "$dir/meta.json" "$other_runner" "$SESSION_ID" <<'PY'
import json, sys
from datetime import datetime, timezone
path, runner, sid = sys.argv[1:4]
now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
meta = {
    "runner": runner,
    "session_id": sid,
    "status": "finished",
    "initial_prompt": "seeded sibling under another runner",
    "created_at": now,
    "updated_at": now,
}
with open(path, "w") as f:
    json.dump(meta, f)
print("seeded", path)
PY

  echo "agent-run sessions after second runner seed:"
  agent-run sessions
  echo
  echo "Note: both rows share session_id=$SESSION_ID but differ by runner."
  echo "Storage paths:"
  echo "  sessions/$RUNNER/$SESSION_ID/"
  echo "  sessions/$other_runner/$SESSION_ID/"
  echo
  echo "Current list format embeds runner in the path: <runner>/<session_id> <status>"
  echo "Proposed unified table (not implemented yet):"
  echo "  SESSION_ID     RUNNER       STATUS"
  echo "  $SESSION_ID    $RUNNER      finished   # (example)"
  echo "  $SESSION_ID    $other_runner finished"
}

main() {
  echo "demo-run-then-sessions: run with --session-id, then list via agent-run sessions"
  build_bins
  run_with_session_id
  list_sessions
  assert_listed
  demo_runner_namespace

  section "done"
  echo "Conclusion: run and sessions share one store (AGENT_RUN_HOME)."
  echo "Sessions are namespaced by runner on disk; list already includes runner as runner/id."
}

main "$@"
