# Scenario

**Feature**: warm Find reuses cache; `.gen` unchanged; meta still correct

```
seed unique grok-tty UUID
  -> Find(UUID)  # warm populate via WarmQueryID
  -> Find(UUID) again
  -> same meta; .gen equal to warm snapshot
```

## Steps

1. Seed one session; set `WarmQueryID` to populate cache.
2. Op `find` same UUID with no mutate.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "33333333-3333-3333-3333-333333333333"
	req.Op = "find"
	req.QueryID = uuid
	req.WarmQueryID = uuid
	req.WarmOp = "find"
	req.Seeds = []SeedMeta{{
		SessionID:       "warm-hit",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	return nil
}
```
