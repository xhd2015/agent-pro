# Scenario

**Feature**: no `mirror sessions: not ready` stderr when grok leaves session without events.jsonl

Reproduces user report after interactive `/exit`:

```
llm-mock run --log-events test.jsonl grok
GROK_HOME=...
llm-mock: mirror sessions: not ready for /Users/.../worktree
```

Exit code is 0, but stderr warning is misleading — there is nothing to mirror when no
`events.jsonl` exists under any session.

## Steps

1. Fake grok on PATH creates session dir without `events.jsonl`, exits 0.
2. `--log-events` set (matches user command).
3. Assert exit 0 and stderr lacks spurious mirror warning.

```go
import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := installFakeGrokEmptySessionExit(t, req); err != nil {
		return err
	}
	req.LogEventsPath = filepath.Join(t.TempDir(), "test.jsonl")
	req.ExecTimeout = 5 * time.Second
	req.ExpectedExit = 0
	return nil
}
```