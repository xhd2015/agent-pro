# Scenario

**Feature**: `resume --detach` without follow-up reopens daemon and prints both ids

```
seed finished bound+exited
  -> agent-run resume --detach <id>
  -> exit 0; stdout session-id + terminal-id; no attach stream noise
```

## Steps

1. Seed bound exited meta (no live registry).
2. Fake TUI hold for reopened daemon.
3. Run `resume --detach <id>` with empty followup.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-resume-detach-reopen"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440801"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-detach-reopen"
	req.InitialPrompt = "prior resume detach"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{"resume", "--detach", req.SessionID}
	req.Mode = "read-meta+registry"
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
