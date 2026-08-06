# Scenario

**Feature**: multi format prints a session header before each session's prompts

```
# two sessions
FormatPromptsListText ->
  ── <idA>  ·  <relative>  ·  <titleA>  ·  <short cwd>
  [ts] promptA
  ── <idB>  ·  ...
  [ts] promptB
```

## Preconditions

- Two sessions with distinct titles and one prompt each.
- Op format-list, !RecentSet, LimitSet with Limit=10 (or default).

## Steps

1. Write two sessions (newest first order known).
2. Format list.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.RecentSet = false
	req.LimitSet = true
	req.Limit = 10

	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatMultiA,
		Title:        "Title Alpha",
		LastActiveAt: atFixed(-5 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("prompt alpha", atFixed(-5*time.Minute))),
	})
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatMultiB,
		Title:        "Title Beta",
		LastActiveAt: atFixed(-15 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("prompt beta", atFixed(-15*time.Minute))),
	})
	return nil
}
```
