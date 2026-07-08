# Scenario

**Feature**: A2 — concurrent EnsureGrokSync yields single active worker

```
pre-seeded updates line on disk
  -> EnsureGrokSync x2 (concurrent)
  -> GrokSyncWorkerCount == 1
  -> one user event for seeded line
```

## Steps

1. Seed one turn in `InitialLines`.
2. Enable `ConcurrentEnsure`.
3. Hold until events flushed.

```go
import "testing"

const idempotentUserPrompt = "idempotent-probe-prompt"

func Setup(t *testing.T, req *Request) error {
	req.InitialLines = []string{
		acpUserMessageChunk(idempotentUserPrompt),
		acpAgentMessageChunk("idempotent-reply"),
		acpTurnCompleted(),
	}
	req.ConcurrentEnsure = true
	req.HoldAfterSchedule = 600 * time.Millisecond
	return nil
}
```