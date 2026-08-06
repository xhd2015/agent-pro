# Scenario

**Feature**: session last_active in window but all prompt timestamps outside → skip

```
# last_active = Now-10m (in window); only user prompt ts = Now-3h (out)
Recent=1h -> session does NOT appear and does NOT consume limit
```

## Preconditions

- Recent=1h, Limit=5 LimitSet=true.
- Target session: last_active recent, prompts all old.
- Companion session with true in-window prompt to prove list still works.

## Steps

1. Write last-active-in / prompts-out session + one real in-window session.
2. List with recent 1h.

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
	req.LimitSet = true
	req.Limit = 5

	// Trap: last_active inside window, prompt ts outside
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idLastActiveOnly,
		Title:        "active but old prompts",
		LastActiveAt: atFixed(-10 * time.Minute),
		Updates: updatesJSONL(
			userChunkAt("ancient prompt", atFixed(-3*time.Hour)),
			assistantChunk("old reply"),
			turnCompleted(),
		),
	})
	// Real in-window session (older last_active than trap so trap would be first if wrongly included)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           multiSessionID(1),
		Title:        "true in-window",
		LastActiveAt: atFixed(-20 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("fresh", atFixed(-20*time.Minute))),
	})
	return nil
}
```
