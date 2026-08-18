# Scenario

**Feature**: Find errors when the provider UUID is absent from store

```
seed unrelated grok-tty with different UUID
  -> FindByGrokSessionID(missing-UUID)
  -> not found; error mentions UUID
```

## Steps

1. Seed one session bound to a different UUID (store non-empty).
2. Query a never-seen UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "find"
	req.QueryID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	req.Seeds = []SeedMeta{{
		SessionID:       "other-sess",
		Runner:          "grok-tty",
		RunnerSessionID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Status:          "finished",
	}}
	return nil
}
```
