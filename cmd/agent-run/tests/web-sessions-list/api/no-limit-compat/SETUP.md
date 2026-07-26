# Scenario

**Feature**: omit `limit` returns all sessions, newest first (backward compat)

```
seed 5 metas (controlled updated_at)
GET /api/agent-run/sessions
  -> sessions length = 5
  -> first session_id is newest (sess-epsilon)
  -> has_more=false when field present
```

## Preconditions

- Five seeded sessions via `defaultFiveSessions()`.
- Expect RED until GET sorts newest-first (today returns unsorted disk order).

## Steps

1. Seed five sessions; start web.
2. GET `/api/agent-run/sessions` with no query params.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "no-limit-compat"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "list", Method: "GET", Path: sessionsPath("")},
	}
	return nil
}
```
