# Scenario

**Feature**: exited session + auto + empty prompt → resume reopen (keep-tty), exit 0

```
seed finished bound+exited
  -> run --auto-send-or-resume --session-id ID --agent-runner-binary REC
  -> exit 0; argv has --resume <id>; empty followup OK
```

## Steps

1. Seed bound exited meta.
2. Install argv runner.
3. Run auto with session-id only (no prompt).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "auto-resume-d2"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440d22"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-auto-resume-d2"
	req.InitialPrompt = "prior turn d2"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvRunner(t, req)
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
