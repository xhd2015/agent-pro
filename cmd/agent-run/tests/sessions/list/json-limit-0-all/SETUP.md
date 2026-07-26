# Scenario

**Feature**: `sessions --json --limit 0` returns all sessions

```
seed 15 -> sessions --json --limit 0 -> array length 15, sorted desc
```

## Preconditions

- Same sort/limit rules as human list.

## Steps

1. Seed 15 sessions.
2. Run `agent-run sessions --json --limit 0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	seedNSessions(t, req.Home, 15)
	req.Args = append(req.Args, "--json", "--limit", "0")
	return nil
}
```
