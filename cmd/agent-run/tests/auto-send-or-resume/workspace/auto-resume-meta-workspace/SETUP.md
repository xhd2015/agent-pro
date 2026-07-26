# Scenario

**Feature**: auto→resume uses `meta.workspace` as child cwd (same rule as resume)

```
meta.workspace=/…/created-ws; CLI cwd=/…/cli-cwd
  -> run --auto-send-or-resume --session-id ID --agent-runner-binary REC "hi"
  -> MODE=resume; child pwd = created-ws
```

## Steps

1. Grouping Setup set Workspace/WorkDir.
2. Seed bound exited meta.
3. Install argv+cwd runner; run auto with followup.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "ws-auto-e2"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440e22"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ws-e2"
	req.InitialPrompt = "created in ws auto"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi from auto e2"
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
