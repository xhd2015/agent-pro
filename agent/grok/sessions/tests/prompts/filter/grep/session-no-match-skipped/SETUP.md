# Scenario

**Feature**: session with prompts but zero grep hits is skipped and does not count toward limit

```
# 3 sessions: match, no-match-only, match; Limit=2, Grep=hit
ListPrompts -> 2 sessions (first and third by last_active); middle skipped
```

## Preconditions

- Three sessions newest-first: A(match), B(no match), C(match).
- LimitSet=2, Grep=hit.
- If B incorrectly counted toward limit, C would be dropped.

## Steps

1. Write A, B, C with distinct last_active.
2. List with grep+limit 2.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.GrepSet = true
	req.Grep = "hit"
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 2

	// A newest
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepA,
		Title:        "A match",
		LastActiveAt: atFixed(-1 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("hit-A", atFixed(-1*time.Minute))),
	})
	// B middle — prompts but no hit
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepB,
		Title:        "B miss",
		LastActiveAt: atFixed(-2 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("only-noise", atFixed(-2*time.Minute))),
	})
	// C oldest of the three — match
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepC,
		Title:        "C match",
		LastActiveAt: atFixed(-3 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("hit-C", atFixed(-3*time.Minute))),
	})
	return nil
}
```
