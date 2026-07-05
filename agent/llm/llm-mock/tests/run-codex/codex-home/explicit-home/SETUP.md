# Scenario

**Feature**: explicit `LLM_MOCK_CODEX_HOME` is used and receives `config.toml`

```
orchestrator -> LLM_MOCK_CODEX_HOME (caller dir)
orchestrator -> write config.toml with base_url -> mock
```

## Steps

1. Create explicit codex home under test temp dir.
2. Set `LLM_MOCK_CODEX_HOME` via `Request.CodexHome`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.CodexHome = filepath.Join(t.TempDir(), "explicit-codex-home")
	return nil
}
```