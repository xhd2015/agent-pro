# Scenario

**Feature**: Find ignores non-grok runners sharing the same provider UUID

```
seed codex-tty + grok-tty with same UUID
  -> FindByGrokSessionID(UUID)
  -> unique grok-tty meta (codex ignored)
```

## Steps

1. Seed `codex-tty` and `grok-tty` with identical `runner_session_id`.
2. Op `find` — only grok-family counts toward cardinality.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	const uuid = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	req.Op = "find"
	req.QueryID = uuid
	req.Seeds = []SeedMeta{
		{SessionID: "codex-hit", Runner: "codex-tty", RunnerSessionID: uuid, Status: "finished"},
		{SessionID: "grok-hit", Runner: "grok-tty", RunnerSessionID: uuid, Status: "finished"},
	}
	return nil
}
```
