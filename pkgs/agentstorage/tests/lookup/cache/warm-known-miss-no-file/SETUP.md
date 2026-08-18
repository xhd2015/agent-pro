# Scenario

**Feature**: after a complete rebuild, unknown UUID is not-found without creating a file

```
seed UUID A; warm Find(A) rebuilds index
  -> Find(never-seen UUID Z)
  -> not-found; no Z.json created
```

## Steps

1. Seed and warm with UUID A so index is complete.
2. Op `find` for a different never-seen UUID Z.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuidA = "44444444-4444-4444-4444-444444444444"
	const uuidZ = "99999999-9999-9999-9999-999999999999"
	req.Op = "find"
	req.QueryID = uuidZ
	req.WarmQueryID = uuidA
	req.WarmOp = "find"
	req.Seeds = []SeedMeta{{
		SessionID:       "known-a",
		Runner:          "grok-tty",
		RunnerSessionID: uuidA,
		Status:          "finished",
	}}
	return nil
}
```
