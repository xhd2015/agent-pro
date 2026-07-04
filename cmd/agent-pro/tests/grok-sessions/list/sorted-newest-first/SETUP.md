# Scenario

**Feature**: list sorts sessions by last_active_at descending

```
# three sessions with non-chronological write order
writeGrokSession(middle, oldest, newest) -> sessions.List

# returned order follows last_active_at not directory creation order
[]Session sorted newest first
```

## Preconditions

- Three sessions have known `last_active_at` timestamps on different hours.

## Steps

1. Write middle timestamp first, then oldest, then newest (out of order).
2. Use `req.Limit = 10` to return all three.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	writeGrokSession(t, req.GrokHome,
		"01900003-0000-7000-8000-000000000002",
		"2026-06-21T12:00:00.000Z", "/tmp/mid", "middle session")
	writeGrokSession(t, req.GrokHome,
		"01900003-0000-7000-8000-000000000001",
		"2026-06-21T10:00:00.000Z", "/tmp/old", "oldest session")
	writeGrokSession(t, req.GrokHome,
		"01900003-0000-7000-8000-000000000003",
		"2026-06-21T14:00:00.000Z", "/tmp/new", "newest session")
	return nil
}
```