# Scenario

**Feature**: grok-sync.json checkpoint enables crash-safe resume without replay

```
worker processes lines -> SaveCheckpoint (offset + turn_index)
  -> StopGrokSync
  -> EnsureGrokSync resumes from checkpoint
  -> new appends only (no turn 1 replay)
```

## Preconditions

- Checkpoint file is separate from `meta.json`.
- Write order: append `events.jsonl` → then `SaveCheckpoint`.

## Steps

1. Grouping leaves configure stop/restart or pre-seeded checkpoint state.
2. Assert checkpoint file contents and event replay semantics.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Runner == "" {
		req.Runner = "grok-tty"
	}
	if req.SessionID == "" {
		req.SessionID = "sync-worker-checkpoint"
	}
	return nil
}
```