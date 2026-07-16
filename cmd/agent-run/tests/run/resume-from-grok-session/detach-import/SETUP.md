# Scenario

**Feature**: import + `--detach` starts keep-alive daemon, prints both ids, and
pre-binds `runner_session_id` (parent exits without waiting for hold sleep)

```
seed Grok UUID; hold fake binary (sleep >> timeout)
  -> agent-run run --detach --session-id FIXED
       --agent-runner-binary HOLD --resume-from-grok-session UUID
  -> exit 0 before hold ends
  -> stdout: session-id: FIXED… / terminal-id: …
  -> meta.runner_session_id = UUID
```

## Preconditions

- Long-hold binary ensures missing `Detach` wiring times out (classic RED).
- Soft grok bind is pre-satisfied via CreateSession pre-bind (P2).
- No `AGENT_RUN_GROK_TTY_COMMAND` (binary path only).

## Steps

1. Seed Grok session at workspace.
2. Install hold runner (120s sleep).
3. Run import with `--detach` and fixed session id; timeout 45s.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "import-detach-s1"
	req.DetachFlag = true
	// holdSec > timeout so headless-without-detach fails by deadline.
	setupDetachOrOpenImport(t, req, 120, 45*time.Second)
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
