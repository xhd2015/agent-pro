# Scenario

**Feature**: unique legacy runner=grok resolve prints bare session id

```
seed legacy-gsid runner=grok runner_session_id=UUID
  -> RunSessions(["resolve", "--grok-session-id", UUID])
  -> stdout "legacy-gsid\n"; err nil
```

## Steps

1. Seed one finished legacy `grok` session.
2. Resolve human mode.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
	req.Seeds = []SeedMeta{{
		SessionID:       "legacy-gsid",
		Runner:          "grok",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	req.Args = []string{"resolve", "--grok-session-id", uuid}
	return nil
}
```
