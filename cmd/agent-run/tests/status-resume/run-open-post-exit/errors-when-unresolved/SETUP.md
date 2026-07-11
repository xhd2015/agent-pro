# Scenario

**Feature**: open post-exit errors when grok session id is not resolved

```
empty GROK_HOME, no session id hook
  -> agent-run run --open "open fail"
  -> (instant attach returns)
  -> exit ≠ 0
  -> stderr: grok session id not resolved …
```

## Steps

1. Point GROK_HOME at empty dir; do not set session id hook.
2. Run open with instant attach + short-hold fake TUI.
3. Assert non-zero exit and not-resolved error.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "empty-grok-home")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	// Explicitly no discovery seed.
	req.GrokSessionUUID = ""
	req.GrokTTYCommand = fakeTUIHoldSeconds(1)
	req.OpenInstantAttach = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", "open fail"}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
