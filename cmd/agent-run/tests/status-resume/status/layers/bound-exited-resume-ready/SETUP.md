# Scenario

**Feature**: bound + exited + dead/absent terminal ⇒ resume ready

```
meta finished, runner_session_id set, no live registry
  -> agent-run status test-exited-s1
  -> runner.exited true, resume.ready yes, process dead / terminal unreachable|missing
```

## Steps

1. Seed finished meta with runner_session_id; no live registry.
2. Run human status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-exited-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440000"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-exited-1"
	req.InitialPrompt = "prior turn"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
