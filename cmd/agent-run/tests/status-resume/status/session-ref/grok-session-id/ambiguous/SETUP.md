# Scenario

**Feature**: two `grok-tty` metas share the same `runner_session_id` → ambiguous

```
seed a1 + a2 both runner=grok-tty runner_session_id=UUID
  -> agent-run status --grok-session-id UUID
  -> exit 1; ambiguous; both agent-run session ids listed
```

## Steps

1. Seed two finished bound sessions with the same provider UUID.
2. Run status with `--grok-session-id` for that UUID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.SessionID = "test-gsid-amb-a"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440904"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-gsid-amb-a"
	req.InitialPrompt = "ambiguous a"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	// Second match (same provider id, different agent-run id).
	seedExtraSessionMeta(t, req,
		"test-gsid-amb-b",
		"grok-tty",
		req.RunnerSessionID,
		"finished",
		"term-gsid-amb-b",
		"ambiguous b",
	)
	req.Args = []string{"status", "--grok-session-id", req.RunnerSessionID}
	return nil
}
```
