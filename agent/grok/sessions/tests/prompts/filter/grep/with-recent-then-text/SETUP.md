# Scenario

**Feature**: recent window applies before grep text filter

```
# one session: old "hit-old" outside window + new "hit-new" + new "miss-new"
Recent=1h Grep=hit -> only hit-new (old hit outside window dropped first)
```

## Preconditions

- RecentSet, Recent=1h, Now=fixedNow.
- hit-old at fixedNow−2h; hit-new and miss-new at fixedNow−10m.
- Op list (or single with list path). Use list with one session.

## Steps

1. Write session with three prompts at different times.
2. List with recent + grep.

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
	req.RecentSet = true
	req.Recent = time.Hour
	req.LimitSet = false // no default cap when RecentSet

	old := atFixed(-2 * time.Hour)
	neu := atFixed(-10 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterGrepA,
		Title:        "recent then grep",
		LastActiveAt: neu,
		Updates: updatesJSONL(
			userChunkAt("hit-old", old),
			turnCompleted(),
			userChunkAt("hit-new", neu),
			turnCompleted(),
			userChunkAt("miss-new", neu.Add(time.Second)),
			turnCompleted(),
		),
	})
	return nil
}
```
