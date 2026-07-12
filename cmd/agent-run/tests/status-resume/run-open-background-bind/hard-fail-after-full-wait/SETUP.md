# Scenario

**Feature**: B3 — bind hard-fails after full wait when discovery is impossible

```
empty GROK_HOME, no session materialization, hard-require path
  -> agent-run run --open "bg bind fail"
  -> (instant attach returns)
  -> wait for bind worker / discovery budget to exhaust
  -> exit ≠ 0
  -> stderr: grok session id not resolved for session …
  -> no false runner_session_id bound
```

## Steps

1. Point GROK_HOME at empty dir; do not set session id hook; never write session files.
2. Run open with instant attach (hard path via GROK_HOME).
3. Assert non-zero exit, not-resolved error, no spurious bind id.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	prompt := "bg bind fail"
	req.OpenPrompt = prompt
	req.InitialPrompt = prompt
	req.GrokHome = filepath.Join(req.TempDir, "empty-grok-home-bg")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	// Explicitly no discovery seed / no session-id hook / no delay schedule.
	req.GrokSessionUUID = ""
	req.GrokMaterializeDelay = 0
	req.GrokTTYCommand = fakeTUIHoldSeconds(2)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", prompt}
	// Full discovery budget may be ~20s+ after implement.
	req.ExecTimeout = 90 * time.Second
	return nil
}
```
