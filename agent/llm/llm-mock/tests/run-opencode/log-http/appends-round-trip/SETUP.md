# Scenario

**Feature**: `--log-http` records HTTP round-trip when fake opencode curls mock once

```
fake opencode -> curl /v1/chat/completions -> log-http JSONL with path + status 200
```

## Steps

1. Config JSON with one exchange for `config-only-prompt` → response `from-config`.
2. Fake opencode curls mock once via baseURL from `$OPENCODE_CONFIG_CONTENT`.
3. `--log-http` points at a fresh `.jsonl` file in temp.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeOpencodeCmd = fakeOpencodeCurlChatCompletionsOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.jsonl")
	return nil
}
```