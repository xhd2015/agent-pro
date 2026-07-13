# Scenario

**Feature**: `status=running` returns only running sessions

```
seed 5 (1 running = sess-beta)
GET ?status=running
  -> sessions: only sess-beta
  -> total=1 when present
```

## Preconditions

- One running session among five.
- `done` semantics covered indirectly via counts leaf (finished+idle).
- Expect RED until status filter is implemented.

## Steps

1. Seed five; start web.
2. GET with `status=running`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "status-running"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "running", Method: "GET", Path: sessionsPath("status=running")},
	}
	return nil
}
```
