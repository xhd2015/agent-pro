# Scenario

**Feature**: `--log-http` rejects paths without `.jsonl` suffix before server listens

```
llm-mock --log-http /tmp/http.log -> startup error (no listener)
```

## Steps

1. Set `LogHTTPFile` to a path ending in `.log` (not `.jsonl`).
2. Do not send HTTP requests — server must fail at startup.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LogHTTPFile = filepath.Join(t.TempDir(), "http.log")
	return nil
}
```