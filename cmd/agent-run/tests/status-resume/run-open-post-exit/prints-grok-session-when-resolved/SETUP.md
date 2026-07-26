# Scenario

**Feature**: open post-exit prints grok session when discovery succeeds and persists meta

```
seed GROK_HOME + AGENT_RUN_GROK_TTY_GROK_SESSION_ID=<uuid>
  -> agent-run run --open "open bind"
  -> (instant attach returns)
  -> stderr: grok-tty: <id>
            grok-tty: grok session <uuid>
            grok-tty: grok updates <path>
  -> meta.runner_session_id == uuid
```

## Steps

1. Seed fake grok session dir with fixed UUID under temp GROK_HOME.
2. Run open with instant attach + respond-hi fake TUI.
3. Mode `read-meta` to load persisted meta after run (session id discovered from
   stderr when possible; also scan home for sessions/grok-tty/*).

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

const openBindGrokUUID = "550e8400-e29b-41d4-a716-446655440700"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "grok-home")
	req.GrokSessionUUID = openBindGrokUUID
	prompt := "open bind"
	req.GrokUpdatesPath = writeFakeGrokSessionDir(t, req.GrokHome, req.TempDir, openBindGrokUUID, prompt)
	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	req.OpenInstantAttach = true
	req.Mode = "read-meta"
	// Session id is assigned by run; leave SessionID empty and resolve in Assert
	// by scanning home or parsing stderr terminal id lines.
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
