# Scenario

**Feature**: RecentSet + LimitSet clips to N in-window sessions

```
# 7 sessions with in-window prompts; Limit=2
ListPrompts -> 2 newest in-window session blocks
```

## Preconditions

- Recent=1h, RecentSet=true, Limit=2, LimitSet=true.
- 7 sessions all with prompts inside the hour.

## Steps

1. Write 7 in-window sessions.
2. List with recent+limit.

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
	req.Limit = 2

	for i := 0; i < 7; i++ {
		id := multiSessionID(i)
		la := atFixed(-time.Duration(i) * time.Minute)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "clip-" + id[len(id)-2:],
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt("p", la)),
		})
	}
	return nil
}
```
