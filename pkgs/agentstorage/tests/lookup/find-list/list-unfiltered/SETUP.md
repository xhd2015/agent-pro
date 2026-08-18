# Scenario

**Feature**: ListBy with no runner filter returns every meta sharing the UUID

```
seed grok-tty + codex-tty same UUID
  -> ListByRunnerSessionID(UUID)  # no runners varargs
  -> len=2 (both runners)
```

## Steps

1. Seed grok and codex with the same provider UUID.
2. Op `list` with empty `Runners` (unfiltered).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	req.Op = "list"
	req.QueryID = uuid
	req.Runners = nil
	req.Seeds = []SeedMeta{
		{SessionID: "list-grok", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
		{SessionID: "list-codex", Runner: "codex-tty", RunnerSessionID: uuid, Status: "finished"},
	}
	return nil
}
```
