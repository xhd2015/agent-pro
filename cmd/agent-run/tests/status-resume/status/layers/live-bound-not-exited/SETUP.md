# Scenario

**Feature**: live bound session with idle/sendable terminal ⇒ not exited, resume not ready

```
meta running + runner_session_id + live registry + idle fake ptywrap
  -> agent-run status test-live-s1
  -> exited: false, resume.ready: no, reason hints send / still active
```

## Steps

1. Start fake ptywrap, write live registry (alive PID), seed meta.
2. Run human status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-live-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440111"
	req.TerminalSessionID = "term-live-1"
	req.InitialPrompt = "live turn"
	seedLiveBoundNotExited(t, req)
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
