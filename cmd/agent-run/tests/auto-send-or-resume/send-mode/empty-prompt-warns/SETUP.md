# Scenario

**Feature**: live session + auto + empty prompt → exit 0 with stderr warning (no send)

```
seed live
  -> agent-run run --auto-send-or-resume --session-id ID
  -> exit 0; stderr warns live / no message to send; no msg_N; no enqueue
```

## Steps

1. Seed live bound session.
2. Run auto with session-id only (no prompt remainder).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "auto-live-empty-c2"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440c22"
	req.TerminalSessionID = "term-auto-live-c2"
	req.InitialPrompt = "prior live empty"
	seedLiveBoundNotExited(t, req)
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
	}
	req.ExecTimeout = 30 * time.Second
	return nil
}
```
