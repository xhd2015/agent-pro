# Scenario

**Feature**: `LLM_MOCK_CONFIG_FILE` resolves config and runs fake grok successfully

```
orchestrator -> mockconfig loader <- LLM_MOCK_CONFIG_FILE
orchestrator -> fake grok (exit 0)
```

## Steps

1. Set `ConfigEnv` to `file` so `Run` exports `LLM_MOCK_CONFIG_FILE`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	return nil
}
```
