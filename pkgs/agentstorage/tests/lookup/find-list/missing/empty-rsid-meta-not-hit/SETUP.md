# Scenario

**Feature**: empty `meta.runner_session_id` never matches a query UUID

```
seed grok-tty with runner_session_id=""
  -> FindByGrokSessionID(some-UUID)
  -> not found (empty rsid is not a hit and not a cache key)
```

## Steps

1. Seed a grok-tty session with empty `RunnerSessionID`.
2. Find a concrete UUID — must miss.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "find"
	req.QueryID = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	req.Seeds = []SeedMeta{{
		SessionID:       "empty-rsid",
		Runner:          "grok-tty",
		RunnerSessionID: "",
		Status:          "finished",
	}}
	return nil
}
```
