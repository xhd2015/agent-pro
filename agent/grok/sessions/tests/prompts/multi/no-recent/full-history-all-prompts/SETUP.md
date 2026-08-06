# Scenario

**Feature**: without --recent, selected sessions include full prompt history

```
# session has prompts at -2h and -10m; !RecentSet
ListPrompts limit 1 -> that session has BOTH prompts (not window-filtered)
```

## Preconditions

- One newest session with two user prompts far apart in time.
- An older decoy session so limit 1 picks the newest only.
- RecentSet=false.

## Steps

1. Write two sessions; newest has prompts at −2h and −10m.
2. List with Limit=1.

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
	req.Limit = 1

	old := atFixed(-2 * time.Hour)
	recent := atFixed(-10 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFullHistory,
		Title:        "full history",
		LastActiveAt: recent,
		Updates: updatesJSONL(
			userChunkAt("old prompt", old),
			assistantChunk("a1"),
			turnCompleted(),
			userChunkAt("recent prompt", recent),
			assistantChunk("a2"),
			turnCompleted(),
		),
	})
	// Older decoy so limit-1 is meaningful
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           multiSessionID(50),
		Title:        "older decoy",
		LastActiveAt: atFixed(-3 * time.Hour),
		Updates:      updatesJSONL(userChunkAt("decoy", atFixed(-3*time.Hour))),
	})
	return nil
}
```
