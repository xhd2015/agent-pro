# Scenario

**Feature**: multi footer counts only real printed prompts, not virtual omission lines

```
# one session, 5 prompts, Head=2 -> footer "... 2 user messages" (not 3, not 5)
```

## Preconditions

- One session 5 prompts; Head=2; Op format-list.
- Footer pattern: `N sessions, M user messages` with M=2.

## Steps

1. Write one session with 5 prompts.
2. Format list with head 2.
3. Assert footer message count.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.HeadSet = true
	req.Head = 2
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 10
	req.ColorMode = "never"
	end := atFixed(-1 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterHead,
		Title:        "footer count",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end, "p1", "p2", "p3", "p4", "p5"),
	})
	return nil
}
```
