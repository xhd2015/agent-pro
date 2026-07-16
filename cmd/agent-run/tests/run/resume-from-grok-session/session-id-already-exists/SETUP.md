# Scenario

**Feature**: `--session-id` that already exists in agent-run store is rejected
for import (distinct from provider **already-mapped** on `runner_session_id`)

```
seed Grok UUID under GROK_HOME
seed agent-run session FIXED (runner=fake-codex; no runner_session_id = UUID)
  -> agent-run run --session-id FIXED --resume-from-grok-session UUID
  -> exit 1; session already exists (or equivalent)
```

## Preconditions

- Pre-existing meta must **not** map the Grok UUID (else already-mapped fires first).
- Grok fixture exists so missing-session does not win first.

## Steps

1. Seed Grok session at workspace.
2. Seed agent-run meta for fixed id without mapping the Grok UUID.
3. Run import with the same `--session-id`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "pre-existing-import-s1"
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	// Collision on agent-run session id only — different runner, no provider map.
	seedExistingAgentSession(t, req, req.SessionID, "fake-codex")
	// No argv runner required: should fail before launch.
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
