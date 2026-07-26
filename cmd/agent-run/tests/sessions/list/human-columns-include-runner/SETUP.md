# Scenario

**Feature**: human list columns include SESSION_ID, RUNNER, STATUS (UPDATED)

```
seed 2 sessions different runners -> sessions --limit 0
-> header SESSION_ID RUNNER STATUS [UPDATED]; bare ids; runner column values
```

## Preconditions

- Human output should include a header row with RUNNER column.
- Session id column is bare id (not runner/id).

## Steps

1. Seed two sessions with known runners and times.
2. Run `agent-run sessions --limit 0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	seedFlatSessionMeta(t, req.Home, "demo_a", "fake-codex", "finished", "2026-07-12T08:00:00Z")
	seedFlatSessionMeta(t, req.Home, "demo_b", "fake-opencode", "running", "2026-07-12T09:00:00Z")
	req.Args = append(req.Args, "--limit", "0")
	return nil
}
```
