# Scenario

**Feature**: `--limit 0` means list all sessions

```
seed 15 -> sessions --limit 0 -> 15 rows, newest first
```

## Preconditions

- Q2: `--limit 0` means all.

## Steps

1. Seed 15 sessions.
2. Run `agent-run sessions --limit 0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	seedNSessions(t, req.Home, 15)
	req.Args = append(req.Args, "--limit", "0")
	return nil
}
```
