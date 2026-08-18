# Scenario

**Feature**: CreateSession with same UUID after unique warm makes next Find ambiguous

```
seed sess1 with UUID; warm Find → unique
  -> CreateSession(sess2, same UUID)  # bumps generation
  -> Find(UUID)
  -> ambiguous; both session ids
```

## Steps

1. Seed unique mapping; warm Find.
2. Mutate `create` second grok-tty with the same UUID.
3. Op `find` — expect ambiguous with asc ids.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "66666666-6666-6666-6666-666666666666"
	req.Op = "find"
	req.QueryID = uuid
	req.WarmQueryID = uuid
	req.WarmOp = "find"
	req.Seeds = []SeedMeta{{
		SessionID:       "first-match",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	req.Mutate = &MutateOp{
		Kind:            "create",
		SessionID:       "second-match",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}
	return nil
}
```
