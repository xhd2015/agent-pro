# Scenario

**Profile**: `real-grok` — production `grok` interactive TUI on PATH (`label: grok`)

```
agent-run run --agent-runner grok-tty (no AGENT_RUN_GROK_TTY_COMMAND)
  -> real grok in PTY
  -> banner detection + prompt injection against live TUI
```

## Preconditions

- Real `grok` CLI must be on `PATH`; otherwise tests skip with `grok not found in PATH`.
- Do **not** set `AGENT_RUN_GROK_TTY_COMMAND` — production grok argv only.
- Leaves require `doctest test --label grok` (excluded from default `./...` runs).
- Longer timeouts acceptable for real LLM TUI startup.

## Steps

1. Grouping `Setup` calls `exec.LookPath("grok")` and sets `req.SkipFakeTUI = true`.
2. Leaf `Setup` sets prompt and optional background attach scenario.
3. `Run` executes against live grok.
4. `Assert` checks banner detection, captured output, or attach connectivity.

```go
import (
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.SkipFakeTUI = true
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_GROK_TTY_COMMAND")
	req.ExecTimeout = 120 * time.Second
	return nil
}
```