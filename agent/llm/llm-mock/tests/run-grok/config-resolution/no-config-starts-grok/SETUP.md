# Scenario

**Feature**: no config env vars → default empty exchanges; fake grok starts successfully

```
orchestrator -> mockconfig loader <- (no config env, default {"exchanges": []})
orchestrator -> llm-mock HTTP server (background)
orchestrator -> fake grok (exit 0)
```

## Steps

1. Omit `LLM_MOCK_CONFIG_FILE`, `LLM_MOCK_CONFIG`, and `--config`.
2. Do not write a config JSON file (`ConfigJSON` empty).
3. Fake grok prints `GROK_HOME=` and exits 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.ExpectedExit = 0
	req.FakeGrokCmd = fakeGrokPrintHome
	return nil
}
```