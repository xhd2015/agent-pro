# Scenario

**Feature**: `agent-run resume` without `--dir` uses `meta.workspace` as child cwd

```
meta.workspace=/…/created-ws; CLI cwd=/…/cli-cwd
  -> agent-run resume --agent-runner-binary REC <id> "hi"
  -> child pwd probe = created-ws (not cli-cwd)
```

## Steps

1. Grouping Setup created created-ws + cli-cwd; Workspace/WorkDir set.
2. Seed bound exited meta with that Workspace.
3. Install argv+cwd runner; run original `resume` (not auto).

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "ws-resume-e1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440e11"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ws-e1"
	req.InitialPrompt = "created in ws"
	req.WriteRegistry = false
	// req.Workspace already set by grouping Setup to created-ws.
	seedBoundExitedDeadTerminal(t, req)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "hi from resume e1"
	req.Args = []string{
		"resume",
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 60 * time.Second
	req.Mode = "read-probes"
	return nil
}
```
