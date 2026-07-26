# Scenario

**Feature**: `q` case-insensitive substring filter on prompt, session_id, workspace, runner

```
seed 5 metas
GET ?q=UNIQUE-QUERY-TOKEN  -> only sess-delta (prompt); total=1
GET ?q=sess-beta           -> only sess-beta (id)
GET ?q=ws-unique           -> only sess-delta (workspace)
GET ?q=opencode            -> only sess-gamma (runner)
GET ?q=UNIQUE-query-token  -> case-insensitive match still sess-delta
```

## Preconditions

- `defaultFiveSessions()` includes UNIQUE-QUERY-TOKEN on delta, opencode runner on gamma.
- Expect RED until `q` is implemented.

## Steps

1. Seed five; start web.
2. Five GET steps covering prompt, id, workspace, runner, and case-insensitivity.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "q-filters"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "q-prompt", Method: "GET", Path: sessionsPath("q=UNIQUE-QUERY-TOKEN")},
		{Name: "q-id", Method: "GET", Path: sessionsPath("q=sess-beta")},
		{Name: "q-workspace", Method: "GET", Path: sessionsPath("q=ws-unique")},
		{Name: "q-runner", Method: "GET", Path: sessionsPath("q=opencode")},
		{Name: "q-case", Method: "GET", Path: sessionsPath("q=UNIQUE-query-token")},
	}
	return nil
}
```
