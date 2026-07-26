# Scenario

**Feature**: default list limit is 10 newest sessions

```
seed 15 sessions with distinct updated_at -> sessions -> 10 rows, newest first
```

## Preconditions

- 15 sessions under flat `sessions/` with `sess_00` oldest … `sess_14` newest.

## Steps

1. Seed 15 sessions via `seedNSessions`.
2. Run `agent-run sessions` (no `--limit`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	seedNSessions(t, req.Home, 15)
	return nil
}
```
