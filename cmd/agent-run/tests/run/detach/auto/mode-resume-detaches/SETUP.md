# Scenario

**Feature**: auto MODE=resume + `--detach` reopens detached (no attach)

```
seed bound+exited
  -> run --auto-send-or-resume --detach --session-id ID
  -> exit 0; both ids; no open/attach; no stream noise
```

## Steps

1. Seed finished bound+exited meta.
2. Fake TUI hold; run auto + detach without followup.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "auto-detach-resume-d1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440d11"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-auto-detach-resume-d1"
	req.InitialPrompt = "prior auto resume detach"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--detach",
		"--session-id", req.SessionID,
		"--agent-runner", "grok-tty",
	}
	req.Mode = "read-meta+registry"
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
