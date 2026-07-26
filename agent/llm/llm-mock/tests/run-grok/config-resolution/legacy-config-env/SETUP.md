# Scenario

**Feature**: legacy `LLM_MOCK_CONFIG` env var still resolves config

```
orchestrator -> mockconfig loader <- LLM_MOCK_CONFIG (legacy fallback)
orchestrator -> fake grok (exit 0)
```

## Steps

1. Set `ConfigEnv` to `legacy` so only `LLM_MOCK_CONFIG` is exported (not `LLM_MOCK_CONFIG_FILE`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "legacy"
	return nil
}
```
