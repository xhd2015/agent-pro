# Scenario

**Feature**: grok leaves session dir without events.jsonl; orchestrator must exit within seconds

Reproduces: `llm-mock run --log-events test.jsonl grok` → open TUI → `/exit` → parent
hangs ~60s (or until Ctrl-C). Interactive grok creates a session tree but may not flush
`events.jsonl` on immediate exit.

```
fake grok (PATH, not hook) -> mkdir GROK_HOME/sessions/<enc-cwd>/<uuid>/ (no events.jsonl) -> exit 0
orchestrator -> must return within ExecTimeout (5s), not poll mirror for 60s
```

## Steps

1. Install fake grok on PATH that creates empty session dir then exits 0.
2. Pass `--log-events` (matches user report).
3. Set `ExecTimeout` to 5 seconds.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if err := installFakeGrokEmptySessionExit(t, req); err != nil {
		return err
	}
	req.LogEventsPath = filepath.Join(t.TempDir(), "test.jsonl")
	req.ExecTimeout = 5 * time.Second
	req.ExpectedExit = 0
	return nil
}
```