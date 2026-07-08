#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PROBE="$REPO_ROOT/script/debug/message-card-ux-probe.js"

if [[ -z "${SCRATCH:-}" ]]; then
  echo "SCRATCH must be set to an output directory" >&2
  exit 1
fi

mkdir -p "$SCRATCH"

if [[ -z "${AGENT_RUN_TOKEN:-}" ]]; then
  echo "AGENT_RUN_TOKEN must be set (from agent-run web startup)" >&2
  exit 1
fi

if ! command -v playwright-debug >/dev/null 2>&1; then
  echo "playwright-debug not found on PATH" >&2
  exit 1
fi

export SCRATCH AGENT_RUN_TOKEN

rm -f "$SCRATCH/message-card-report.json" \
  "$SCRATCH/message-card-report-1.json" \
  "$SCRATCH/message-card-report-2.json"

export PROBE_RUN=1
playwright-debug run "$PROBE" | tee "$SCRATCH/probe-run1.log"

export PROBE_RUN=2
playwright-debug run "$PROBE" | tee "$SCRATCH/probe-run2.log"

python3 - <<'PY'
import json, os, sys
scratch = os.environ["SCRATCH"]
path = os.path.join(scratch, "message-card-report.json")
if not os.path.isfile(path) or os.path.getsize(path) == 0:
    print(f"missing or empty report: {path}", file=sys.stderr)
    sys.exit(1)
with open(path) as f:
    report = json.load(f)
runs = report.get("runs")
if not isinstance(runs, list) or len(runs) != 2:
    print(f"expected 2 runs in report, got {runs!r}", file=sys.stderr)
    sys.exit(1)
if not report.get("pass"):
    print(f"report pass=false: {json.dumps(report, indent=2)}", file=sys.stderr)
    sys.exit(1)
for run in runs:
    if run.get("issues"):
        print(f"run {run.get('run')} has issues: {run['issues']}", file=sys.stderr)
        sys.exit(1)
print("message-card probe twice: OK")
PY