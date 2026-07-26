# Scenario

**Feature**: real codex run produces visible scrollback output

```
real codex PTY
  -> banner detected
  -> prompt submitted
  -> stdout or events contain visible assistant/codex text
```

## Preconditions

- Real `codex` on PATH (`t.Skip` if absent).
- No fake TUI.

## Steps

1. Run `agent-run run --agent-runner codex-tty "say hi"`.
2. Assert the run exits successfully and captured output is non-empty.

```go
import (
	"os/exec"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.SkipFakeTUI = true
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND")
	req.Args = []string{"run", "--agent-runner", "codex-tty", "say hi"}
	req.ExecTimeout = 120 * time.Second
	return nil
}
```
