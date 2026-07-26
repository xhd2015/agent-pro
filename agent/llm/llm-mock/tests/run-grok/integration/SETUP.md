# Scenario

**Profile**: `real-grok, slow` — production grok headless with mocked LLM backend

```
orchestrator -> llm-mock HTTP server (background)
orchestrator -> isolated GROK_HOME + config.toml
real grok -p "..." -m mock-model -> mock LLM
integration <- stdout Paris + events.jsonl turn_started mock-model
```

## Preconditions

- Real `grok` CLI must be on `PATH`; otherwise tests skip with `grok not found in PATH`.
- Do **not** set `LLM_MOCK_RUN_GROK_COMMAND` — production grok argv only.
- Leaves require `doctest test --label real-grok` (excluded from default runs).
- Longer timeout acceptable for real LLM session startup.

## Steps

1. Grouping `Setup` calls `exec.LookPath("grok")` and skips if missing.
2. Leaf `Setup` sets headless grok args and mock config exchange.
3. `Run` executes against live grok with mock backend.
4. `Assert` checks Paris in output and `turn_started` with `model_id: mock-model`.

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
	req.SkipRealGrok = true
	req.FakeGrokCmd = ""
	req.ConfigEnv = "file"
	req.ExecTimeout = 120 * time.Second
	req.ExpectParis = true
	req.ExpectMockModel = true
	return nil
}
```
