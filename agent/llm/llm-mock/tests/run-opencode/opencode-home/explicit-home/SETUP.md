# Scenario

**Feature**: explicit `LLM_MOCK_OPENCODE_CONFIG_DIR` and `LLM_MOCK_OPENCODE_HOME` are used

```
orchestrator -> LLM_MOCK_OPENCODE_CONFIG_DIR + LLM_MOCK_OPENCODE_HOME (caller dirs)
orchestrator -> OPENCODE_CONFIG_CONTENT with baseURL -> mock
```

## Steps

1. Create explicit opencode home and config dir under test temp dir.
2. Set `LLM_MOCK_OPENCODE_HOME` and `LLM_MOCK_OPENCODE_CONFIG_DIR` via `Request`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	base := t.TempDir()
	req.OpencodeHome = filepath.Join(base, "explicit-opencode-home")
	req.OpencodeConfigDir = filepath.Join(base, "explicit-opencode-config")
	return nil
}
```