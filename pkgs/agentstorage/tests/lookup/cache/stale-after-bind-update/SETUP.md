# Scenario

**Feature**: UpdateSessionRunnerSessionID bumps store gen; next Find sees new mapping

```
seed sess unbound (empty rsid); warm Find(targetUUID) → not-found, builds empty-ish index
  -> UpdateSessionRunnerSessionID(sess, targetUUID)  # bumps generation
  -> Find(targetUUID)
  -> returns sess (stale cache not trusted)
```

## Steps

1. Seed a grok-tty session with empty `runner_session_id`.
2. Warm Find of the eventual UUID (miss; rebuilds index).
3. Mutate `update_rsid` to bind the UUID.
4. Op `find` same UUID — must resolve after gen bump + rebuild.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "55555555-5555-5555-5555-555555555555"
	req.Op = "find"
	req.QueryID = uuid
	req.WarmQueryID = uuid
	req.WarmOp = "find"
	req.Seeds = []SeedMeta{{
		SessionID:       "bind-later",
		Runner:          "grok-tty",
		RunnerSessionID: "",
		Status:          "finished",
	}}
	req.Mutate = &MutateOp{
		Kind:            "update_rsid",
		SessionID:       "bind-later",
		RunnerSessionID: uuid,
	}
	return nil
}
```
