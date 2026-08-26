# Scenario

**Feature**: multi list keeps only matching prompts inside each session block

```
# two sessions with mixed match/non-match prompts
ListPrompts(Grep=keepme) -> each block only keepme lines
```

## Preconditions

- Two sessions (A newer than B); each has one match and one non-match.
- Grep=`keepme`, !RecentSet, LimitSet high enough for both.
- Op `list`.

## Steps

1. Write two sessions.
2. List with grep.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "list"
	req.GrepSet = true
	req.Grep = []string{"keepme"}
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 10

	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepA,
		Title:        "grep multi A",
		LastActiveAt: atFixed(-2 * time.Minute),
		Updates: chronoPromptUpdates(atFixed(-2*time.Minute),
			"keepme-alpha",
			"ignore-other",
		),
	})
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepB,
		Title:        "grep multi B",
		LastActiveAt: atFixed(-10 * time.Minute),
		Updates: chronoPromptUpdates(atFixed(-10*time.Minute),
			"noise",
			"keepme-beta",
		),
	})
	return nil
}
```
