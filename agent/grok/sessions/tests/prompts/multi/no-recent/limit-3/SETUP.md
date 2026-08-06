# Scenario

**Feature**: LimitSet with N=3 returns three newest sessions

```
# !RecentSet LimitSet Limit=3; 8 sessions on disk
ListPrompts -> 3 newest SessionPrompts
```

## Preconditions

- 8 sessions; Limit=3, LimitSet=true.

## Steps

1. Write 8 sessions.
2. Call ListPrompts with Limit 3.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 3
	for i := 0; i < 8; i++ {
		id := multiSessionID(i)
		la := atFixed(-time.Duration(i) * time.Minute)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "lim3-" + id[len(id)-2:],
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt("p", la)),
		})
	}
	return nil
}
```
