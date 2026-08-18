# Scenario

**Feature**: two grok-tty metas sharing one UUID → Find ambiguous; List returns both

```
seed sess-b then sess-a (same UUID, runner=grok-tty)
  -> FindByGrokSessionID(UUID)
  -> ambiguous; session ids "sess-a, sess-b" (asc)
  -> ListByRunnerSessionID(UUID, "grok", "grok-tty") len=2
```

## Steps

1. Seed two sessions in reverse session_id order so asc sorting is load-bearing.
2. Op `find_and_list` with grok runner filter.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	req.Op = "find_and_list"
	req.QueryID = uuid
	req.Runners = []string{"grok", "grok-tty"}
	req.Seeds = []SeedMeta{
		{SessionID: "sess-b", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
		{SessionID: "sess-a", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
	}
	return nil
}
```
