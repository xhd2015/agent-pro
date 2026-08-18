# Scenario

**Feature**: unique legacy `runner=grok` session found by provider UUID

```
seed session legacy-gsid runner=grok runner_session_id=UUID
  -> FindByGrokSessionID(UUID)
  -> Meta.SessionID == legacy-gsid
```

## Steps

1. Seed one finished legacy `grok` session (not `grok-tty`).
2. Op `find` with that UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	req.Op = "find"
	req.QueryID = uuid
	req.Seeds = []SeedMeta{{
		SessionID:       "legacy-gsid",
		Runner:          "grok",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	return nil
}
```
