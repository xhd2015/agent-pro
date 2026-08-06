# Scenario

**Feature**: RecentSet + Limit when only one session matches

```
# 1 in-window session; several outside; Limit=5
ListPrompts -> exactly 1 block
```

## Preconditions

- Recent=30m, Limit=5 LimitSet=true.
- One session with prompt 10m ago.
- Two sessions with prompts 2h ago only.

## Steps

1. Write 1 in-window + 2 outside sessions.
2. List with recent 30m and limit 5.

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
	req.LimitSet = true
	req.Limit = 5

	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           multiSessionID(0),
		Title:        "only-in",
		LastActiveAt: atFixed(-10 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("inside", atFixed(-10*time.Minute))),
	})
	for i := 1; i <= 2; i++ {
		id := multiSessionID(i)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "out-" + id[len(id)-2:],
			LastActiveAt: atFixed(-2 * time.Hour),
			Updates:      updatesJSONL(userChunkAt("outside", atFixed(-2*time.Hour))),
		})
	}
	return nil
}
```
