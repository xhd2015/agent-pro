#!/usr/bin/env bash
# Mark every running session in AGENT_RUN_HOME as finished (for real filter-empty probe).
set -euo pipefail

AGENT_RUN_HOME="${AGENT_RUN_HOME:-}"
if [[ -z "${AGENT_RUN_HOME}" ]]; then
  echo "AGENT_RUN_HOME must be set" >&2
  exit 1
fi

python3 - "${AGENT_RUN_HOME}" <<'PY'
import json, sys
from datetime import datetime, timezone
from pathlib import Path

home = Path(sys.argv[1])
count = 0
now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
for meta_path in home.glob("sessions/*/*/meta.json"):
    with open(meta_path) as f:
        meta = json.load(f)
    if meta.get("status") != "running":
        continue
    meta["status"] = "finished"
    meta["updated_at"] = now
    with open(meta_path, "w") as f:
        json.dump(meta, f, separators=(",", ":"))
    count += 1
print(f"finished {count} running sessions under {home}")
if count == 0:
    raise SystemExit("no running sessions to finish")
PY