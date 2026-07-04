# Scenario

**Feature**: explicit `LLM_MOCK_GROK_HOME` is used and receives `config.toml`

```
orchestrator -> LLM_MOCK_GROK_HOME (caller dir)
orchestrator -> write config.toml with models_base_url -> mock
```

## Steps

1. Create explicit grok home under test temp dir.
2. Set `LLM_MOCK_GROK_HOME` via `Request.GrokHome`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(t.TempDir(), "explicit-grok-home")
	return nil
}
```
