# Scenario

**Profile**: `mock-ui` — real Codex interactive TUI with `llm-mock-run-codex`
(mock LLM only). Label: `codex`.

```
agent-run run --agent-runner codex-tty --agent-runner-binary llm-mock-run-codex --open
  -> real Codex UI + mock Responses API
```

## Preconditions

- `codex` on PATH (else skip).
- Build `llm-mock-run-codex` **and sibling** `llm-mock` into the same bin dir.
- Do **not** set `AGENT_RUN_CODEX_TTY_COMMAND` (real UI).
- `LLM_MOCK_CONFIG_FILE` points at exchange JSON.
- Leaves require `doctest test --label codex` (excluded from default unlabeled runs).

## Steps

1. Skip if `codex` missing.
2. `SkipFakeTUI`; clear fake TTY env.
3. Build mock binaries; write default mock config.
4. Leaf sets `Mode` (`mock-ui-open-idle` / `mock-ui-send`) and prompts.

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
	req.UseLLMMockCodex = true
	req.ExecTimeout = 120 * time.Second
	ensureLLMMockCodex(t, req)
	writeDefaultMockCodexConfig(t, req)
	return nil
}
```
