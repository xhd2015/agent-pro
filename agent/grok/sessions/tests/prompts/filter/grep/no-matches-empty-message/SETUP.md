# Scenario

**Feature**: when no prompts survive grep, format shows friendly empty message

```
# session has prompts; Grep never matches
format-list -> "No user prompts found\n"
```

## Preconditions

- One session with non-matching prompts.
- Grep=`zzzz-no-hit`, Op format-list.

## Steps

1. Write session.
2. Format list with grep that matches nothing.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.GrepSet = true
	req.Grep = []string{"zzzz-no-hit"}
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 10
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepA,
		Title:        "empty after grep",
		LastActiveAt: atFixed(-5 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("hello world", atFixed(-5*time.Minute))),
	})
	return nil
}
```
