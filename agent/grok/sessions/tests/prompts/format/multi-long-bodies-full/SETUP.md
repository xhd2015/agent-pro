# Scenario

**Feature**: multi-session format prints each long prompt body in full by default

```
# two sessions, each with 220-x user prompt; !MaxBodySet
FormatPromptsListText -> both full 220-x bodies; no body-cap …
```

## Preconditions

- Two sessions (idFormatMultiA/B) with long bodies, distinct last_active for order.
- Op format-list; LimitSet if needed so both appear (or only two sessions total).
- No MaxBody.

## Steps

1. Write two sessions with long prompts.
2. List + format without MaxBody.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-list"
	req.LimitSet = true
	req.Limit = 10
	// Newest first: A more recent than B.
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatMultiA,
		Title:        "long-A",
		LastActiveAt: atFixed(-1 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("A-"+longPromptRunes(220), atFixed(-1*time.Minute))),
	})
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatMultiB,
		Title:        "long-B",
		LastActiveAt: atFixed(-2 * time.Minute),
		Updates:      updatesJSONL(userChunkAt("B-"+longPromptRunes(220), atFixed(-2*time.Minute))),
	})
	return nil
}
```
