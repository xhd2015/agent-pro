# Scenario

**Feature**: `--log-events` rejects paths without `.jsonl` suffix before codex starts

```
llm-mock run --log-events /tmp/events.log codex
CLI validation error (.jsonl) -> no codex, no log file
```

## Steps

1. Set `LogEventsPath` to a non-`.jsonl` path.
2. Fake codex hook would print `CODEX_RAN` if codex started — must not appear.
3. Expect non-zero exit.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.log")
	req.FakeCodexCmd = `sh -c 'echo CODEX_RAN; exit 0'`
	req.ExpectedExit = 1
	return nil
}
```