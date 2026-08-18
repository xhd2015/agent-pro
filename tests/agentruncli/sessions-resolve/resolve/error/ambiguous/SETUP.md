# Scenario

**Feature**: two grok-tty metas with same UUID → ambiguous resolve error

```
seed sess-b then sess-a (same UUID)
  -> RunSessions resolve --grok-session-id UUID
  -> ambiguous; ids "sess-a, sess-b" (asc); stdout empty
```

## Steps

1. Seed two sessions in reverse session_id order so asc sort is load-bearing.
2. Resolve that UUID.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	req.Seeds = []SeedMeta{
		{SessionID: "sess-b", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
		{SessionID: "sess-a", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
	}
	req.Args = []string{"resolve", "--grok-session-id", uuid}
	return nil
}
```
