# Scenario

**Feature**: list table output shows expected columns and relative last-active times

```
# three sessions with last_active_at offsets from fixed now
writeGrokSession x3 -> sessions.List -> FormatListTable(now)

# table includes SESSION ID, LAST ACTIVE, TITLE, CWD with relative deltas
terminal table text
```

## Preconditions

- `req.Now` is fixed at `2026-07-03T15:00:00.000Z` by root Setup.
- Table format is the default list output.

## Steps

1. Create session A active at `req.Now` → `just now`.
2. Create session B active 5 minutes before `req.Now` → `5m ago`.
3. Create session C active 2 hours before `req.Now` → `2h ago`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	now := req.Now.UTC()

	writeGrokSession(t, req.GrokHome,
		"01900004-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		now.Format(time.RFC3339), "/tmp/project-a", "Alpha refactor")
	writeGrokSession(t, req.GrokHome,
		"01900004-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		now.Add(-5*time.Minute).Format(time.RFC3339), "/tmp/project-b", "Beta bugfix")
	writeGrokSession(t, req.GrokHome,
		"01900004-cccc-7ccc-cccc-cccccccccccc",
		now.Add(-2*time.Hour).Format(time.RFC3339), "/tmp/project-c", "Gamma cleanup")
	return nil
}
```