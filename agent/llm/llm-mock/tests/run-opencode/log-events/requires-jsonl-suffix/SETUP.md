# Scenario

**Feature**: `--log-events` rejects paths without `.jsonl` suffix before opencode starts

```
llm-mock run --log-events /tmp/events.log opencode
CLI validation error (.jsonl) -> no opencode, no log file
```

## Steps

1. Set `LogEventsPath` to a non-`.jsonl` path.
2. Fake opencode hook would print `OPENCODE_RAN` if opencode started — must not appear.
3. Expect non-zero exit.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.log")
	req.FakeOpencodeCmd = `sh -c 'echo OPENCODE_RAN; exit 0'`
	req.ExpectedExit = 1
	return nil
}
```