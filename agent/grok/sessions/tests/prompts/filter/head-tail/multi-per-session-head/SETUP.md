# Scenario

**Feature**: multi list applies head per session independently with own M

```
# two sessions × 3 prompts each; Head=1
format-list -> each block: first prompt + (...2 omitted...)
```

## Preconditions

- Sessions A (newer) and B; each has pA1,pA2,pA3 / pB1,pB2,pB3.
- Head=1; Op format-list; Limit high.

## Steps

1. Write two sessions with 3 prompts each.
2. Format list with head 1.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.HeadSet = true
	req.Head = 1
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 10
	req.ColorMode = "never"

	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepA,
		Title:        "multi head A",
		LastActiveAt: atFixed(-2 * time.Minute),
		Updates: chronoPromptUpdates(atFixed(-2*time.Minute), "a1", "a2", "a3"),
	})
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepB,
		Title:        "multi head B",
		LastActiveAt: atFixed(-8 * time.Minute),
		Updates: chronoPromptUpdates(atFixed(-8*time.Minute), "b1", "b2", "b3"),
	})
	return nil
}
```
