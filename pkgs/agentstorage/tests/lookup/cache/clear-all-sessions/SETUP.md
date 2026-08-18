# Scenario

**Feature**: ClearAllSessions drops index mapping; subsequent Find is not-found

```
seed + warm Find → cache populated
  -> ClearAllSessions
  -> index/by-runner-session gone or empty; Find(UUID) not-found
```

## Steps

1. Seed and warm to populate cache.
2. Mutate `clear_all`.
3. Op `find` same UUID — not found; mapping dir gone or without UUID files.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "77777777-7777-7777-7777-777777777777"
	req.Op = "find"
	req.QueryID = uuid
	req.WarmQueryID = uuid
	req.WarmOp = "find"
	req.Seeds = []SeedMeta{{
		SessionID:       "to-clear",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	req.Mutate = &MutateOp{Kind: "clear_all"}
	return nil
}
```
