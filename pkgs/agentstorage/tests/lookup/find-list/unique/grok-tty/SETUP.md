# Scenario

**Feature**: unique `grok-tty` session found by provider UUID

```
seed session hello-gsid runner=grok-tty runner_session_id=UUID
  -> FindByGrokSessionID(UUID)
  -> Meta.SessionID == hello-gsid
```

## Steps

1. Seed one finished `grok-tty` session with a fixed UUID.
2. Op `find` with that UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	req.Op = "find"
	req.QueryID = uuid
	req.Seeds = []SeedMeta{{
		SessionID:       "hello-gsid",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	return nil
}
```
