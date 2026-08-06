# Scenario

**Feature**: no flags → default limit of 10 newest sessions

```
# 15 sessions on disk; !RecentSet !LimitSet
ListPrompts -> 10 SessionPrompts (newest by last_active)
```

## Preconditions

- 15 sessions with last_active_at = fixedNow − i minutes (i=0..14).
- Each has one user prompt.
- LimitSet=false, Limit=0.

## Steps

1. Write 15 sessions via multiSessionID.
2. Call ListPrompts with defaults.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.RecentSet = false
	req.LimitSet = false
	req.Limit = 0
	// id 0 newest (last_active = fixedNow), id 14 oldest
	for i := 0; i < 15; i++ {
		id := multiSessionID(i)
		la := atFixed(-time.Duration(i) * time.Minute)
		writePromptSession(t, req.GrokHome, promptSessionOpts{
			ID:           id,
			Title:        "sess-" + id[len(id)-2:],
			LastActiveAt: la,
			Updates:      updatesJSONL(userChunkAt("p-"+id[len(id)-2:], la)),
		})
	}
	return nil
}
```

