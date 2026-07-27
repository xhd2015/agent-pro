# Scenario

**Bug**: real codex receives prompt chars in the input buffer but never gets Enter;
`run ls` sits in the prompt without executing.

```
agent-run run --agent-runner codex-tty "run ls" → codex should execute ls and
scrollback should show directory listing (not only echoed prompt text)
```

## Preconditions

- Real `codex` on PATH (`t.Skip` if absent).
- No `AGENT_RUN_CODEX_TTY_COMMAND` override.

## Steps

1. Run with prompt `run ls` against real codex interactive TUI.

```go
import (
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.SkipFakeTUI = true
	req.Args = []string{"run", "--agent-runner", "codex-tty", "run ls"}
	req.ExecTimeout = 120 * time.Second
	return nil
}
```