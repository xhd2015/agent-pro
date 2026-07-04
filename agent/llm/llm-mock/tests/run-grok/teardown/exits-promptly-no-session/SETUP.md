# Scenario

**Feature**: grok exits immediately with no session tree; orchestrator returns promptly

Regression guard: no unnecessary mirror polling when `sessions/` does not exist.

## Steps

1. Fake grok on PATH: `exit 0` only (no session dirs).
2. `--log-events` set like user command.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	if err := installFakeGrokImmediateExit(t, req); err != nil {
		return err
	}
	req.LogEventsPath = filepath.Join(t.TempDir(), "test.jsonl")
	req.ExecTimeout = 5 * time.Second
	req.ExpectedExit = 0
	return nil
}
```