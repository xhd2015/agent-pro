# Scenario

**Feature**: live session + auto + `--open` is rejected (open only for run/resume)

```
seed live
  -> agent-run run --auto-send-or-resume --open --session-id ID "hi"
  -> exit 1; --open not valid while live
```

## Steps

1. Seed live bound session.
2. Run auto with `--open` and a prompt.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "auto-live-open-c3"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440c33"
	req.TerminalSessionID = "term-auto-live-c3"
	req.InitialPrompt = "prior live open"
	req.FollowupPrompt = "should not open-send"
	seedLiveBoundNotExited(t, req)
	req.OpenInstantAttach = true // even with instant attach, live auto must reject --open
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--open",
		"--session-id", req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 30 * time.Second
	return nil
}
```
