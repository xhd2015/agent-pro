# Scenario

**Feature**: `status --grok-session-id` also matches exact `meta.runner=grok`
(not only `grok-tty`)

```
seed runner=grok runner_session_id=UUID finished
  -> agent-run status --grok-session-id UUID
  -> exit 0; resolves that session
```

## Steps

1. Seed finished bound meta with `runner=grok` and known UUID.
2. Run status with `--grok-session-id` only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok"
	req.SessionID = "test-gsid-grok-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440902"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-gsid-grok-1"
	req.InitialPrompt = "grok runner resolve"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{"status", "--grok-session-id", req.RunnerSessionID}
	return nil
}
```
