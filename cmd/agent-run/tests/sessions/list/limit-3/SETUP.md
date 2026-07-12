# Scenario

**Feature**: `--limit 3` returns three newest sessions

```
seed 15 -> sessions --limit 3 -> sess_14, sess_13, sess_12
```

## Preconditions

- Same 15-session fixture as default-limit leaf.

## Steps

1. Seed 15 sessions.
2. Run `agent-run sessions --limit 3`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	seedNSessions(t, req.Home, 15)
	req.Args = append(req.Args, "--limit", "3")
	return nil
}
```
