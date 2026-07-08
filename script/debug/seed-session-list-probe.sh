#!/usr/bin/env bash
# Seed mixed running + finished sessions for session-list UX probe.
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8192}"
AGENT_RUN_HOME="${AGENT_RUN_HOME:-}"

echo "Seeding session-list probe fixtures against ${BASE_URL}"

# Ensure at least one finished session with a readable prompt preview.
FINISHED_META=$(find "${AGENT_RUN_HOME}/sessions" -name meta.json 2>/dev/null | head -1 || true)
if [[ -n "${FINISHED_META}" ]]; then
  python3 - "$FINISHED_META" <<'PY'
import json, sys
from datetime import datetime, timezone
path = sys.argv[1]
with open(path) as f:
    meta = json.load(f)
meta["status"] = "finished"
meta["updated_at"] = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
with open(path, "w") as f:
    json.dump(meta, f, separators=(",", ":"))
print("patched finished:", path)
PY
fi

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