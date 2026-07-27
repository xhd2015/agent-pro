# Scenario

**Profile**: `real-codex` — production `codex` interactive TUI on PATH (`label: codex`)

```
agent-run run --agent-runner codex-tty (no AGENT_RUN_CODEX_TTY_COMMAND)
  -> real codex in PTY
  -> banner detection + prompt injection against live TUI
```

## Preconditions

- Real `codex` CLI must be on `PATH`; otherwise tests skip with `codex not found in PATH`.
- Do **not** set `AGENT_RUN_CODEX_TTY_COMMAND` — production codex argv only.
- Leaves require `doctest test --label codex` (excluded from default `./...` runs).
- Longer timeouts acceptable for real LLM TUI startup.

## Steps

1. Grouping `Setup` calls `exec.LookPath("codex")` and sets `req.SkipFakeTUI = true`.
2. Leaf `Setup` sets prompt and optional background attach scenario.
3. `Run` executes against live codex.
4. `Assert` checks banner detection, captured output, or attach connectivity.

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
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND")
	req.ExecTimeout = 120 * time.Second
	return nil
}
```