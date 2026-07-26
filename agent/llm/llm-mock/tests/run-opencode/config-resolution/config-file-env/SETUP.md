# Scenario

**Feature**: `LLM_MOCK_CONFIG_FILE` with Paris exchange resolves config and runs fake opencode successfully

```
orchestrator -> mockconfig loader <- LLM_MOCK_CONFIG_FILE (Paris exchange)
orchestrator -> fake opencode (exit 0)
```

## Steps

1. Set `ConfigEnv` to `file` so `Run` exports `LLM_MOCK_CONFIG_FILE`.
2. Use Paris exchange JSON per integration default.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.ConfigJSON = `{
  "port": 8080,
  "exchanges": [
    {
      "request": {"content": "capital of France", "index": -1},
      "response": {"content": "The capital of France is Paris.", "finish_reason": "stop"}
    }
  ]
}`
	return nil
}
```