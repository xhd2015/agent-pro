# Scenario

**Feature**: B1 — bind completes before/with open when GROK session is preseeded

```
seed GROK_HOME + AGENT_RUN_GROK_TTY_GROK_SESSION_ID=<uuid>
  -> agent-run run --open "bg bind preseed"
  -> (instant attach returns; bind worker finds preseeded session)
  -> stderr: grok-tty: grok session <uuid>
  -> meta.runner_session_id == uuid; exit 0
```

## Steps

1. Pre-seed fake grok session dir with fixed UUID under temp GROK_HOME.
2. Run open with instant attach + short-hold fake TUI.
3. Assert stderr session lines and persisted `runner_session_id`.

```go
import (
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const bgBindPreseedUUID = "550e8400-e29b-41d4-a716-446655440801"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prompt := "bg bind preseed"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = bgBindPreseedUUID
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, bgBindPreseedUUID, prompt)
	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
