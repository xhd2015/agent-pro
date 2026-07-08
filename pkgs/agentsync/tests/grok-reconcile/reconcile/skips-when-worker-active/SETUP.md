# Scenario

**Feature**: R3 — reconcile skips when in-process worker already active

```
EnsureGrokSync worker running and syncing
  -> ReconcileOnce on same session
  -> no duplicate event lines appended
```

## Steps

1. Seed running session with grok updates + known grok session id.
2. Start `EnsureGrokSync`; wait for initial sync.
3. Call `ReconcileOnce`; assert event count unchanged.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "reconcile-skip-worker"
	req.SessionID = "reconcile-skip-test"
	req.InitialPrompt = reconcileSkipPrompt
	req.GrokSessionID = reconcileSkipGrokUUID
	req.WorkerHold = 600 * time.Millisecond
	return nil
}
```