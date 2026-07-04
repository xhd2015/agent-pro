# Scenario

**Feature**: `--log-http` records HTTP round-trip when fake grok curls mock once

```
fake grok -> curl /v1/chat/completions -> log-http JSONL with path + status 200
```

## Steps

1. Config JSON with one exchange for `config-only-prompt` → response `from-config`.
2. Fake grok curls mock once via `GROK_MODELS_BASE_URL`.
3. `--log-http` points at a fresh `.jsonl` file in temp.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeGrokCmd = fakeGrokCurlOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.LogHTTPPath = filepath.Join(t.TempDir(), "session-log-http.jsonl")
	return nil
}
```