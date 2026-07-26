# Scenario

**Feature**: no config env vars → default empty exchanges; fake opencode starts successfully

```
orchestrator -> mockconfig loader <- (no config env, default {"exchanges": []})
orchestrator -> llm-mock HTTP server (background)
orchestrator -> fake opencode (exit 0)
```

## Steps

1. Omit `LLM_MOCK_CONFIG_FILE` and `--config`.
2. Do not write a config JSON file (`ConfigJSON` empty).
3. Fake opencode prints `OPENCODE_CONFIG_DIR=` and exits 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigJSON = ""
	req.ConfigEnv = ""
	req.ExpectedExit = 0
	req.FakeOpencodeCmd = fakeOpencodePrintConfigDir
	return nil
}
```