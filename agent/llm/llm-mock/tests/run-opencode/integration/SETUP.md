# Scenario

**Profile**: `real-opencode, slow` — production opencode headless with mocked LLM backend

```
orchestrator -> llm-mock HTTP server (background)
orchestrator -> isolated HOME + OPENCODE_CONFIG_DIR + OPENCODE_CONFIG_CONTENT (openai-compatible)
real opencode run "What is the capital of France?" --model llm-mock/mock-model -> mock LLM
integration <- output Paris; HTTP log has chat/completions with mock model
```

## Preconditions

- Real `opencode` CLI must be on `PATH`; otherwise tests skip with `opencode not found in PATH`.
- Do **not** set `LLM_MOCK_RUN_OPENCODE_COMMAND` — production opencode argv only.
- Leaves require `doctest test --label real-opencode` (excluded from default runs).
- Longer timeout acceptable for real opencode session startup (60s per requirement).
- Exit 0 is **not** required — agent loop may continue after configured exchanges exhaust.

## Steps

1. Grouping `Setup` calls `exec.LookPath("opencode")` and skips if missing.
2. Leaf `Setup` sets headless opencode args, Paris mock config exchange, and `--log-http` for request proof.
3. `Run` executes against live opencode with mock backend.
4. `Assert` checks Paris in output and mock HTTP log model/path.

```go
import (
	"os/exec"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if _, err := exec.LookPath("opencode"); err != nil {
		t.Skip("opencode not found in PATH")
	}
	req.SkipRealOpencode = true
	req.FakeOpencodeCmd = ""
	req.ConfigEnv = "file"
	req.ExecTimeout = 60 * time.Second
	req.ExpectParis = true
	return nil
}
```