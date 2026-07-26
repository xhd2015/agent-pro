# Scenario

**Bug**: `run --auto-send-or-resume` (MODE=resume) must error clearly when `meta.workspace` is gone (no `--dir`)

```
meta.workspace=/…/gone-ws (missing); bound+exited; no --dir
  -> run --auto-send-or-resume --session-id ID --agent-runner-binary REC "followup"
  -> MODE=resume; exit 1; stderr: session workspace missing + path + --dir hint
  -> must NOT only look like fork/exec binary-missing
```

## Steps

1. Grouping Setup set cli-cwd; override Workspace to a never-created path under TempDir.
2. Seed bound+exited meta so auto classifies MODE=resume.
3. Install argv+cwd runner; run auto path with followup, no `--dir`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	gone := filepath.Join(req.TempDir, "gone-ws")
	_ = os.RemoveAll(gone)
	req.Workspace = gone

	req.SessionID = "ws-auto-missing-e5"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440e55"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ws-e5"
	req.InitialPrompt = "created then workspace deleted auto"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "followup auto after gone workspace"
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
