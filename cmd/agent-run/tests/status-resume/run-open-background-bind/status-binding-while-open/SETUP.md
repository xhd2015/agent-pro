# Scenario

**Feature**: B4 — concurrent `status` while open bind is in flight shows binding or bound

```
fixed --session + empty GROK_HOME + delayed materialization (~2.5s)
  + AGENT_RUN_OPEN_ATTACH_INSTANT=1
  -> start agent-run run --open in background
  -> wait until session meta exists
  -> agent-run status --json <session>  (mid-open probe)
  -> runner.status is "binding" or "bound" (not idle unbound without bind signal)
  -> open completes: meta.runner_session_id set; exit 0
```

## Steps

1. Use fixed `--session` so status can target the open session.
2. Schedule delayed GROK materialization so bind is still pending shortly after meta appears.
3. Mode `open-status-mid`: run open async, probe status, wait for open.
4. Assert mid-open runner status and final bind success.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

const bgBindStatusUUID = "550e8400-e29b-41d4-a716-446655440804"
const bgBindStatusSession = "open-bg-bind-status-1"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prompt := "bg bind status probe"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt
	req.SessionID = bgBindStatusSession
	req.GrokHome = filepath.Join(req.TempDir, "grok-home-status")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.GrokSessionUUID = bgBindStatusUUID
	req.GrokMaterializeDelay = 2500 * time.Millisecond
	req.GrokTTYCommand = fakeTUIHoldSeconds(10)
	req.OpenInstantAttach = true
	req.Mode = "open-status-mid"
	// Probe after meta + bind.json in_progress; CI parallel load can exceed 300ms.
	req.StatusProbeAfter = 1500 * time.Millisecond
	req.Args = []string{
		"run", "--agent-runner", "grok-tty",
		"--session", bgBindStatusSession,
		"--open", prompt,
	}
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
