# Scenario

**Feature**: live session + auto + `--open` is accepted (open ignored; send proceeds)

```
seed live
  -> agent-run run --auto-send-or-resume --open --session-id ID "hi"
  -> exit 0; stdout msg_N; send path (not resume)
```

## Steps

1. Seed live bound session.
2. Run auto with `--open` and a prompt.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "auto-live-open-c3"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440c33"
	req.TerminalSessionID = "term-auto-live-c3"
	req.InitialPrompt = "prior live open"
	req.FollowupPrompt = "live open send followup"
	seedLiveBoundNotExited(t, req)
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-live-open-should-not-exist.log")
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--open",
		"--session-id", req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
