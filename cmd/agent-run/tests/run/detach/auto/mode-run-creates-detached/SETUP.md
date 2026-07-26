# Scenario

**Feature**: auto MODE=run + `--detach` creates a detached keep-alive session

```
no meta for id
  -> run --auto-send-or-resume --detach --session-id NEW "prompt"
  -> exit 0; both ids; registry keep-alive; no attach stream
```

## Steps

1. Do **not** seed meta.
2. Fake TUI hold for daemon.
3. Run auto + detach with new session-id and prompt.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "auto-detach-new-r1"
	req.Prompt = "create detached please"
	req.WorkDir = req.TempDir
	req.Workspace = req.TempDir
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--detach",
		"--session-id", req.SessionID,
		"--agent-runner", "grok-tty",
		req.Prompt,
	}
	req.Mode = "read-meta+registry"
	req.ExecTimeout = 60 * time.Second
	if fileExists(metaJSONPath(req.Home, req.SessionID)) {
		t.Fatalf("precondition: meta must not exist: %s", metaJSONPath(req.Home, req.SessionID))
	}
	return nil
}
```
