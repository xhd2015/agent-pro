# Scenario

**Feature**: `--log-http` rejects paths without `.jsonl` suffix before codex starts

```
llm-mock run --log-http /tmp/http.log codex
CLI validation error (.jsonl) -> no codex, no log file
```

## Steps

1. Set `LogHTTPPath` to a non-`.jsonl` path.
2. Fake codex hook would print `CODEX_RAN` if codex started — must not appear.
3. Expect non-zero exit.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.log")
	req.FakeCodexCmd = `sh -c 'echo CODEX_RAN; exit 0'`
	req.ExpectedExit = 1
	return nil
}
```