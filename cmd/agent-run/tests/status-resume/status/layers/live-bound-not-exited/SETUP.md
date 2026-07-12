# Scenario

**Feature**: live bound session with idle/sendable terminal ⇒ not exited, resume not ready (control for zombie-serve)

```
meta running + runner_session_id + live registry + idle fake ptywrap (Grok ›)
  -> agent-run status test-live-s1
  -> exited: false, resume.ready: no, reason hints send / still active
```

## Steps

1. Start fake ptywrap with idle/sendable scrollback, write live registry (alive PID), seed meta.
2. Run human status.
3. Regression: must stay exited false when truly live (E2/E4 vs zombie-serve E1).

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
