# Scenario

**Feature**: `--log-events` rejects paths without `.jsonl` suffix before grok starts

```
llm-mock run --log-events /tmp/events.log grok
CLI validation error (.jsonl) -> no grok, no log file
```

## Steps

1. Set `LogEventsPath` to a non-`.jsonl` path.
2. Fake grok hook would print `GROK_RAN` if grok started — must not appear.
3. Expect non-zero exit.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.LogEventsPath = filepath.Join(t.TempDir(), "events.log")
	req.FakeGrokCmd = `sh -c 'echo GROK_RAN; exit 0'`
	req.ExpectedExit = 1
	return nil
}
```