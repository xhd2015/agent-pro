#!/usr/bin/env bash
# Seed mixed running + finished sessions for session-list UX probe.
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8192}"
AGENT_RUN_HOME="${AGENT_RUN_HOME:-}"

if [[ -z "${AGENT_RUN_HOME}" ]]; then
  echo "AGENT_RUN_HOME must be set to the server's home directory" >&2
  exit 1
fi

echo "Seeding session-list probe fixtures against ${BASE_URL}"
echo "AGENT_RUN_HOME=${AGENT_RUN_HOME}"

# Ensure at least one finished session with a readable prompt preview.
ensure_finished_session() {
  local finished_dir="${AGENT_RUN_HOME}/sessions/opencode/web_finished_probe001"
  if [[ -d "${finished_dir}" ]]; then
    python3 - "${finished_dir}/meta.json" <<'PY'
import json, sys
from datetime import datetime, timezone
path = sys.argv[1]
with open(path) as f:
    meta = json.load(f)
meta["status"] = "finished"
meta["updated_at"] = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
with open(path, "w") as f:
    json.dump(meta, f, separators=(",", ":"))
print("refreshed finished:", path)
PY
    return
  fi
  mkdir -p "${finished_dir}"
  python3 - "${finished_dir}/meta.json" <<'PY'
import json, sys
from datetime import datetime, timezone
path = sys.argv[1]
now = datetime.now(timezone.utc)
meta = {
    "runner": "opencode",
    "session_id": "web_finished_probe001",
    "initial_prompt": "Refactor the home session list for better agent chat UX",
    "status": "finished",
    "workspace": "/Users/xhd2015/.wrk/worktrees/agent-pro-master-2026-07-06-agent-run-web-use-same-tty-watch-attach",
    "created_at": (now.replace(hour=now.hour - 2)).strftime("%Y-%m-%dT%H:%M:%SZ"),
    "updated_at": now.strftime("%Y-%m-%dT%H:%M:%SZ"),
}
with open(path, "w") as f:
    json.dump(meta, f, separators=(",", ":"))
open(path.replace("meta.json", "events.jsonl"), "a").close()
print("created finished:", path)
PY
}
ensure_finished_session

seed_scroll_sessions() {
  local count="${SEED_SCROLL_COUNT:-20}"
  local runner="fake-codex"
  local base_epoch
  base_epoch=$(python3 -c 'import time; print(int(time.time()) - 3600)')
  local i=1
  while [[ "${i}" -le "${count}" ]]; do
    local session_id
    session_id=$(printf 'home-sess-%03d' "${i}")
    local sess_dir="${AGENT_RUN_HOME}/sessions/${runner}/${session_id}"
    if [[ ! -d "${sess_dir}" ]]; then
      mkdir -p "${sess_dir}"
      local updated_at created_at ts
      updated_at=$(python3 -c "import datetime; print((datetime.datetime.fromtimestamp(${base_epoch} + ${i} * 60, datetime.timezone.utc)).strftime('%Y-%m-%dT%H:%M:%SZ'))")
      created_at=$(python3 -c "import datetime; print((datetime.datetime.fromtimestamp(${base_epoch} + ${i} * 60 - 120, datetime.timezone.utc)).strftime('%Y-%m-%dT%H:%M:%SZ'))")
      ts=$(python3 -c "print((${base_epoch} + ${i} * 60) * 1000)")
      python3 - "${sess_dir}/meta.json" "${runner}" "${session_id}" "${created_at}" "${updated_at}" <<'PY'
import json, sys
path, runner, session_id, created_at, updated_at = sys.argv[1:6]
meta = {
    "runner": runner,
    "session_id": session_id,
    "initial_prompt": f"Home scroll seed session {session_id}",
    "status": "idle",
    "workspace": "/tmp/home-scroll-workspace",
    "created_at": created_at,
    "updated_at": updated_at,
}
with open(path, "w") as f:
    json.dump(meta, f, separators=(",", ":"))
PY
      printf '{"type":"message","role":"user","text":"Home scroll seed session %s","timestamp":%s}\n' "${session_id}" "${ts}" > "${sess_dir}/events.jsonl"
    fi
    i=$((i + 1))
  done
  echo "seeded ${count} scroll sessions under ${runner}"
}
seed_scroll_sessions

# Create a live running session via API.
RUNNING_JSON=$(curl -sf -X POST "${BASE_URL}/api/agent-run/sessions" \
  -H 'Content-Type: application/json' \
  -d '{"runner":"opencode","prompt":"Resume this running agent chat about session list UX"}')
echo "created running session: ${RUNNING_JSON}"

curl -sf "${BASE_URL}/api/agent-run/sessions" | python3 -c '
import json, sys
data = json.load(sys.stdin)
statuses = [s.get("status") for s in data.get("sessions", [])]
print("session statuses:", statuses)
if "running" not in statuses:
    raise SystemExit("no running session after seed")
if "finished" not in statuses:
    raise SystemExit("no finished session after seed")
'