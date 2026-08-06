# Scenario

**Feature**: all prompts outside the recent window → empty multi list

```
# Recent=30m; only sessions with prompts ≥2h old
ListPrompts -> empty slice
```

## Preconditions

- RecentSet=true, Recent=30m, LimitSet=false.
- Two sessions with last_active and prompt times 2h ago.

## Steps

1. Write two outside-window sessions.
2. Call ListPrompts.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.RecentSet = true
	req.Recent = 30 * time.Minute
	req.LimitSet = false

	for i := 0; i < 2; i++ {
		id := multiSessionID(i)
		ts := atFixed(-2 * time.Hour)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "old-" + id[len(id)-2:],
			LastActiveAt: ts,
			Updates:      updatesJSONL(userChunkAt("stale", ts)),
		})
	}
	return nil
}
```
