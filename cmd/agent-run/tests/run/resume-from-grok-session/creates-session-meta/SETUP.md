# Scenario

**Feature**: successful import creates agent-run meta with `runner=grok-tty` and
pre-bound `runner_session_id` (Grok UUID); workspace from Grok cwd

```
seed Grok UUID (info.cwd = WorkDir); no mapping
argv-recorder binary (so launch can complete)
  -> agent-run run --session-id FIXED --agent-runner-binary REC
       --resume-from-grok-session UUID "meta followup"
  -> exit 0
  -> sessions/FIXED/meta.json: runner=grok-tty, runner_session_id=UUID
```

## Steps

1. Seed Grok session at process workspace.
2. Install argv-recorder (launch path; not asserted here for argv content).
3. Run with fixed `--session-id` so meta path is known.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "import-meta-s1"
	req.FollowupPrompt = "create meta followup"
	setupValidImport(t, req, true)
	req.Args = runArgs(req, req.GrokSessionID, req.FollowupPrompt)
	return nil
}
```
