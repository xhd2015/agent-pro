# Scenario

**Feature**: `--agent-runner grok` (non-TTY) is rejected for
`--resume-from-grok-session` — only exact `grok-tty` is allowed when the runner
flag is set

```
seed Grok session UUID under GROK_HOME
  -> agent-run run --agent-runner grok --resume-from-grok-session UUID
  -> exit 1; requires grok-tty
```

## Steps

1. Seed Grok session at process workspace.
2. Run with `--agent-runner grok` (distinct from `grok-tty`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokCWD = absPath(t, req.WorkDir)
	seedGrokSession(t, req.GrokHome, req.GrokCWD, req.GrokSessionID)
	req.AgentRunner = "grok"
	req.Args = runArgs(req, req.GrokSessionID)
	return nil
}
```
