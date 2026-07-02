# Scenario

**Bug**: real grok receives prompt chars in the input buffer but never gets Enter;
`run ls` sits in the prompt without executing.

```
agent-run run --agent-runner grok-tty "run ls" → grok should execute ls and
scrollback should show directory listing (not only echoed prompt text)
```

## Preconditions

- Real `grok` on PATH (`t.Skip` if absent).
- No `AGENT_RUN_GROK_TTY_COMMAND` override.

## Steps

1. Run with prompt `run ls` against real grok interactive TUI.

```go
import (
	"os/exec"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.SkipFakeTUI = true
	req.Args = []string{"run", "--agent-runner", "grok-tty", "run ls"}
	req.ExecTimeout = 120 * time.Second
	return nil
}
```