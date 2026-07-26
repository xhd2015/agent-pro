# Scenario

**Profile**: `real-codex, slow` — production codex headless with mocked LLM backend

```
orchestrator -> llm-mock HTTP server (background)
orchestrator -> isolated CODEX_HOME + config.toml (wire_api=responses)
real codex exec -m mock-model "What is the capital of France?" -> mock LLM
integration <- stdout Paris within 60s
```

## Preconditions

- Real `codex` CLI must be on `PATH`; otherwise tests skip with `codex not found in PATH`.
- Do **not** set `LLM_MOCK_RUN_CODEX_COMMAND` — production codex argv only.
- Leaves require `doctest test --label real-codex` (excluded from default runs).
- Longer timeout acceptable for real codex session startup (60s per requirement).

## Steps

1. Grouping `Setup` calls `exec.LookPath("codex")` and skips if missing.
2. Leaf `Setup` sets headless codex args and mock config exchange.
3. `Run` executes against live codex with mock backend.
4. `Assert` checks Paris in output.

```go
import (
	"os/exec"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not found in PATH")
	}
	req.SkipRealCodex = true
	req.FakeCodexCmd = ""
	req.ConfigEnv = "file"
	req.ExecTimeout = 60 * time.Second
	req.ExpectParis = true
	return nil
}
```