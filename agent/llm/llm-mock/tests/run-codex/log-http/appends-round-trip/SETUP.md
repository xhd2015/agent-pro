# Scenario

**Feature**: `--log-http` records HTTP round-trip when fake codex curls mock once

```
fake codex -> curl /v1/responses -> log-http JSONL with path + status 200
```

## Steps

1. Config JSON with one exchange for `config-only-prompt` → response `from-config`.
2. Fake codex curls mock once via base URL from `$CODEX_HOME/config.toml`.
3. `--log-http` points at a fresh `.jsonl` file in temp.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ConfigEnv = "file"
	req.FakeCodexCmd = fakeCodexCurlResponsesOnce
	req.ConfigJSON = minimalMockConfigJSON(8080, "")
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.jsonl")
	return nil
}
```