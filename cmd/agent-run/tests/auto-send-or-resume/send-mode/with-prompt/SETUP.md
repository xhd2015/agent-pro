# Scenario

**Feature**: live session + auto + non-empty prompt → send (msg_N), not resume

```
seed live sendable terminal
  -> agent-run run --auto-send-or-resume --session-id ID "followup"
  -> exit 0; stdout msg_N; inject/enqueue to terminal_session_id
  -> must NOT spawn provider with --resume
```

## Steps

1. Seed live bound not-exited session (fake ptywrap inject APIs + registry).
2. Reserve an argv probe path that must remain absent (send does not spawn runner).
3. Run auto with prompt.

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "auto-live-send-c1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440c11"
	req.TerminalSessionID = "term-auto-live-c1"
	req.InitialPrompt = "prior live"
	req.FollowupPrompt = "auto send followup c1"
	seedLiveBoundNotExited(t, req)
	// Argv probe path reserved: must remain absent (send path does not spawn runner).
	req.ArgvProbePath = filepath.Join(req.TempDir, "argv-probe-should-not-exist.log")
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"--session-id", req.SessionID,
		req.FollowupPrompt,
	}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
