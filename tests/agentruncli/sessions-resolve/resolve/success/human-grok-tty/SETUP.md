# Scenario

**Feature**: unique grok-tty resolve prints bare session id

```
seed hello-world runner=grok-tty runner_session_id=UUID
  -> RunSessions(["resolve", "--grok-session-id", UUID])
  -> stdout "hello-world\n"; err nil
```

## Steps

1. Seed one finished `grok-tty` session.
2. Resolve with that UUID (human mode).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
	req.Seeds = []SeedMeta{{
		SessionID:       "hello-world",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	req.Args = []string{"resolve", "--grok-session-id", uuid}
	return nil
}
```
