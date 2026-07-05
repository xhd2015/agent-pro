# Scenario

**Feature**: `--log-http` rejects paths without `.jsonl` suffix before opencode starts

```
llm-mock run --log-http /tmp/http.log opencode
CLI validation error (.jsonl) -> no opencode, no log file
```

## Steps

1. Set `LogHTTPPath` to a non-`.jsonl` path.
2. Fake opencode hook would print `OPENCODE_RAN` if opencode started — must not appear.
3. Expect non-zero exit.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LogHTTPPath = filepath.Join(t.TempDir(), "http.log")
	req.FakeOpencodeCmd = `sh -c 'echo OPENCODE_RAN; exit 0'`
	req.ExpectedExit = 1
	return nil
}
```