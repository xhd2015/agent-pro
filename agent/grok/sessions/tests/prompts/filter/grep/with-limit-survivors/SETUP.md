# Scenario

**Feature**: with grep, --limit counts only sessions that still have ≥1 match

```
# 5 sessions: match, miss, match, miss, match (newest first)
Grep=hit Limit=2 -> only first two matching sessions (positions 0 and 2)
```

## Preconditions

- Five sessions i=0..4; match when i even (`hit-N`), miss when i odd.
- Limit=2, Grep=hit.
- Op list.

## Steps

1. Write five sessions.
2. List with grep+limit.

```go
import (
	"fmt"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.GrepSet = true
	req.Grep = []string{"hit"}
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 2

	for i := 0; i < 5; i++ {
		id := multiSessionID(i)
		la := atFixed(-time.Duration(i) * time.Minute)
		text := fmt.Sprintf("miss-%d", i)
		if i%2 == 0 {
			text = fmt.Sprintf("hit-%d", i)
		}
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        fmt.Sprintf("lim-surv-%d", i),
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt(text, la)),
		})
	}
	return nil
}
```
