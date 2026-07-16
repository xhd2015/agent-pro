# Scenario

**Feature**: registry remains alive after `resume --detach`

```
seed bound+exited
  -> agent-run resume --detach <id>
  -> exit 0
  -> registry for printed terminal-id exists + TCP reachable
```

## Steps

1. Seed exited bound meta.
2. Run resume detach with hold TUI; assert registry keep-alive.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "test-resume-detach-alive"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440802"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-resume-detach-alive"
	req.InitialPrompt = "prior resume detach alive"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	setGrokTTYCommand(req, fakeTUIHoldSeconds(45))
	req.Args = []string{"resume", "--detach", req.SessionID}
	req.Mode = "detach-registry-after"
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
