# Scenario

**Feature**: `--resume-from-grok-session` is mutually exclusive with
`--auto-send-or-resume`

```
agent-run run --auto-send-or-resume --session-id X
              --agent-runner-binary REC
              --resume-from-grok-session UUID
  -> exit ≠ 0
  -> error mentions mutually exclusive / cannot use both
```

## Preconditions

- Grok fixture + argv-recorder so a missing-mutex implementation finishes the
  import path quickly (exit 0) rather than hanging on a real TTY — still RED
  because we require non-zero exit + exclusive wording.
- Prefer early flag validation once implemented (no launch).

## Steps

1. Seed Grok session.
2. Install argv-recorder (keeps wrong-path launch finite).
3. Run both flags with `--session-id` (needed if auto path wins).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "mutex-auto-s1"
	req.AutoSendOrResume = true
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	installArgvRunner(t, req)
	req.ExecTimeout = 45 * time.Second
	req.Args = runArgs(req, req.GrokSessionID, "mutex prompt unused")
	return nil
}
```
