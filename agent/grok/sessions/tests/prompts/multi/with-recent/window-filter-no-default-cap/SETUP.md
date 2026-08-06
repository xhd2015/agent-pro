# Scenario

**Feature**: RecentSet without LimitSet returns all in-window sessions (no default 10)

```
# Recent=1h; 12 sessions each with one prompt inside the hour
ListPrompts -> 12 SessionPrompts (not clipped to 10)
```

## Preconditions

- RecentSet=true, Recent=1h, LimitSet=false.
- 12 sessions with last_active and prompt ts within the last hour.
- Plus 1 session with only a prompt older than 1h (must be excluded).

## Steps

1. Write 12 in-window sessions + 1 out-of-window.
2. List with Recent 1h only.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.RecentSet = true
	req.Recent = time.Hour
	req.LimitSet = false
	req.Limit = 0

	for i := 0; i < 12; i++ {
		id := multiSessionID(i)
		// Stagger within the hour: 0..11 minutes ago
		la := atFixed(-time.Duration(i) * time.Minute)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "win-" + id[len(id)-2:],
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt("in-window", la)),
		})
	}
	// Out of window: prompt 3h ago; last_active also old
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           multiSessionID(90),
		Title:        "outside",
		LastActiveAt: atFixed(-3 * time.Hour),
		Updates:      updatesJSONL(userChunkAt("too old", atFixed(-3*time.Hour))),
	})
	return nil
}
```
