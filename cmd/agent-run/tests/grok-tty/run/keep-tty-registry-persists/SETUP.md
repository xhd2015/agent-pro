# Scenario

**Feature**: `--keep-tty` flag preserves registry entry after run completes

```
agent-run run --agent-runner grok-tty --keep-tty "hi"
  -> stderr grok-tty: session-N
  -> prompts injected, response captured
  -> run exits normally (turn complete)
  -> registry entry PERSISTS (not removed)
```

## Preconditions

- Fake TUI (`AGENT_RUN_GROK_TTY_COMMAND`) responds quickly so the run exits promptly.
- Without `--keep-tty`, the registry file is removed after run exits.

## Steps

1. Set `req.Args` to include `--keep-tty` and a simple prompt.
2. Use the default `fakeTUIRespondHi()` which outputs "Response: hi" and exits.
3. After `Run` completes, verify the registry JSON file still exists at its expected path.

```go
import (
	"testing"
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--keep-tty", "hi"}
	req.GrokTTYCommand = fakeTUIRespondHi()
	setGrokTTYCommand(req, req.GrokTTYCommand)

	t.Cleanup(func() {
		regPath := filepath.Join(grokTTYRegistryDir(req.Home), req.GrokTTYSessionID+".json")
		_ = os.Remove(regPath)
	})
	return nil
}
```
