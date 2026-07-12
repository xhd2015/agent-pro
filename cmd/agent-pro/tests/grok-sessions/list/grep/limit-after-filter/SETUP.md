# Scenario

**Feature**: --limit applies after grep filter to the newest matching sessions

```
# 5 sessions: newest has no match; next four match GREP_LIMIT_TOKEN with decreasing ages
writeGrokSession x5 -> ListWithGrep(limit=2)

# result: exactly 2 sessions — the two newest matches (not overall newest)
```

## Preconditions

- Limit must not count non-matching sessions.
- Sort among matches is still newest `last_active_at` first.
- Non-matching newest session must not appear in the result.

## Steps

1. Set `req.Grep = "GREP_LIMIT_TOKEN"`, `req.Limit = 2`, `req.Color = "never"`.
2. Write sessions at T+4 (newest, no match), T+3..T+0 (match, older).

```go
import (
	"fmt"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Grep = "GREP_LIMIT_TOKEN"
	req.Limit = 2
	req.Color = "never"

	now := req.Now.UTC()

	// Newest overall — does NOT match.
	writeGrokSession(t, req.GrokHome,
		"01900016-0000-7000-8000-000000000005",
		now.Add(-1*time.Minute).Format(time.RFC3339),
		"/tmp/grep-limit-nomatch",
		"Newest but no token here")

	// Four matching sessions, newest match first in timestamps.
	// IDs 004 newest match ... 001 oldest match.
	for i := 4; i >= 1; i-- {
		id := fmt.Sprintf("01900016-0000-7000-8000-%012d", i)
		// i=4 → 2m ago, i=3 → 3m ago, ... i=1 → 5m ago
		ts := now.Add(-time.Duration(6-i) * time.Minute).Format(time.RFC3339)
		summaryPath := writeGrokSession(t, req.GrokHome,
			id,
			ts,
			fmt.Sprintf("/tmp/grep-limit-%d", i),
			fmt.Sprintf("Match work %d GREP_LIMIT_TOKEN", i))
		writeChatHistory(t, sessionDirOf(summaryPath), []chatHistoryMsg{
			{Type: "user", Content: fmt.Sprintf("note %d about GREP_LIMIT_TOKEN", i)},
		})
	}
	return nil
}
```
