# Scenario

**Feature**: --json resolve prints session meta object

```
seed hello-json runner=grok-tty runner_session_id=UUID status=finished
  -> RunSessions(["resolve", "--json", "--grok-session-id", UUID])
  -> JSON {session_id, runner, runner_session_id, status}; err nil
```

## Steps

1. Seed one finished grok-tty session.
2. Resolve with `--json`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"
	req.Seeds = []SeedMeta{{
		SessionID:       "hello-json",
		Runner:          "grok-tty",
		RunnerSessionID: uuid,
		Status:          "finished",
	}}
	req.Args = []string{"resolve", "--json", "--grok-session-id", uuid}
	return nil
}
```
