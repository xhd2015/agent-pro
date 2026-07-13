# Scenario

**Bug**: `agent-run resume` must error clearly when `meta.workspace` no longer exists (no `--dir`)

```
meta.workspace=/…/gone-ws (missing); CLI cwd exists; no --dir
  -> agent-run resume --agent-runner-binary REC <id> "followup"
  -> exit 1; stderr: session workspace missing + path + --dir hint
  -> must NOT only look like fork/exec binary-missing
```

## Steps

1. Grouping Setup created created-ws + cli-cwd; override Workspace to a never-created path under TempDir.
2. Seed bound+exited meta with that missing Workspace (WriteRegistry false).
3. Install argv+cwd runner (would record if spawn happened; must not be needed for happy path).
4. Run original `resume` without `--dir`.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	gone := filepath.Join(req.TempDir, "gone-ws")
	// Ensure the path does not exist (never create; remove if leftover).
	_ = os.RemoveAll(gone)
	req.Workspace = gone

	req.SessionID = "ws-missing-e4"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440e44"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ws-e4"
	req.InitialPrompt = "created then workspace deleted"
	req.WriteRegistry = false
	seedBoundExitedDeadTerminal(t, req)

	installArgvCwdRunner(t, req)
	req.FollowupPrompt = "followup after gone workspace"
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
