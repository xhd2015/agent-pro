# Scenario

**Feature**: real `llm-mock run codex` must not emit codex model-metadata warning for `mock-model`

Reproduces user report after `llm-mock run --mock-events-preset=think-tool-message codex`:

```
warning: Model metadata for `mock-model` not found. Defaulting to fallback metadata; this can degrade performance and cause issues.
```

## Preconditions

- Real `codex` on PATH.
- No `LLM_MOCK_RUN_CODEX_COMMAND` hook.

## Steps

1. Run headless codex with think-tool-message preset (user example).
2. Assert mocked assistant text appears and metadata warning is absent.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SkipRealCodex = true
	req.FakeCodexCmd = ""
	req.ConfigEnv = ""
	req.MockEventsPreset = "think-tool-message"
	req.ExecTimeout = 60 * time.Second
	req.CodexArgs = []string{
		"exec",
		"--skip-git-repo-check",
		"-m", "mock-model",
		"hello",
	}
	return nil
}
```