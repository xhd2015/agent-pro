# Scenario

**Feature**: auto + brand-new `--session-id` + prompt takes MODE=run (create session)

```
no meta for id
  -> agent-run run --auto-send-or-resume --session-id NEW --agent-runner-binary RECORDER "prompt"
  -> exit 0; meta created; argv has prompt; NO --resume; meta.workspace set
```

## Steps

1. Do **not** seed meta.
2. Install argv-recording fake runner (no `AGENT_RUN_GROK_TTY_COMMAND`).
3. Run auto with new session-id and prompt; Mode=read-meta.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "auto-new-session-b1"
	req.FollowupPrompt = "create me please"
	// Ensure create-time workspace is TempDir (CLI WorkDir).
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir
	installArgvRunner(t, req)
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		"--agent-runner-binary", req.AgentRunnerBinary,
		req.FollowupPrompt,
	}
	req.Mode = "read-meta"
	req.ExecTimeout = 60 * time.Second
	// Sanity: meta must not pre-exist.
	if fileExists(metaJSONPath(req.Home, defaultRunner, req.SessionID)) {
		t.Fatalf("precondition: meta must not exist: %s", filepath.Join(req.Home, "sessions", defaultRunner, req.SessionID))
	}
	return nil
}
```
