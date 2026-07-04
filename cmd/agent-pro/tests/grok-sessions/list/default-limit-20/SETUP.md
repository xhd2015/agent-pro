# Scenario

**Feature**: default list limit returns 20 newest of many sessions

```
# 25 summary.json files with ascending last_active_at
writeGrokSession x25 -> sessions.List(grokHome, 20)

# only the 20 most recent sessions are returned
[]Session (len=20) -> newest timestamps kept
```

## Preconditions

- Default limit is 20 when `req.Limit` is zero.
- Twenty-five distinct session directories exist.

## Steps

1. Create 25 sessions with `last_active_at` from `2026-06-23T00:00:00.000Z`
   through `2026-06-23T00:24:00.000Z` (one per minute).
2. Leave `req.Limit` at zero to exercise the default.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	for i := 0; i < 25; i++ {
		id := fmt.Sprintf("01900001-0000-7000-8000-%012d", i+1)
		ts := fmt.Sprintf("2026-06-23T00:%02d:00.000Z", i)
		writeGrokSession(t, req.GrokHome, id, ts, "/tmp/project", fmt.Sprintf("session %d", i+1))
	}
	return nil
}
```