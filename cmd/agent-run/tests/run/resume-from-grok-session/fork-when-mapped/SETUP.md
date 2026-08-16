# Scenario

**Feature**: `--fork` with already-mapped Grok id succeeds and records
`grok --resume <uuid> --fork-session` argv

```
seed Grok UUID under GROK_HOME
seed agent-run meta mapped to UUID (already mapped)
argv-recorder as --agent-runner-binary
  -> agent-run run --session-id NEW --fork
       --resume-from-grok-session UUID "followup"
  -> exit 0
  -> ARGV_RECORD has --resume UUID and --fork-session
  -> NEW meta exists; mapped parent meta unchanged
```

## Steps

1. Seed Grok session + already-mapped meta.
2. Install argv recorder; set fixed new session id + ForkFlag.
3. Run import with --fork.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	req.MappedSessID = "mapped-parent-fork"
	seedMappedMeta(t, req, "grok-tty", req.MappedSessID, req.GrokSessionID)
	req.SessionID = "fork-child-sess-1"
	req.FollowupPrompt = "fork followup"
	req.ForkFlag = true
	installArgvRunner(t, req)
	if req.ExecTimeout < 60*time.Second {
		req.ExecTimeout = 60 * time.Second
	}
	req.Args = runArgs(req, req.GrokSessionID, req.FollowupPrompt)
	return nil
}
```
