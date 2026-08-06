# Scenario

**Feature**: exclude drops matching prompt lines; non-matches remain

```
# prompts: keep-this, drop-noise, also-keep
Exclude="noise" -> keep-this, also-keep
```

## Preconditions

- Single session; ExcludeSet, Exclude=`noise`.
- Op single.

## Steps

1. Write three prompts.
2. Filter with exclude noise.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "single"
	req.SessionID = idFilterSingle
	req.ExcludeSet = true
	req.Exclude = "noise"
	end := atFixed(-4 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "exclude drops",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			"keep-this",
			"drop-noise",
			"also-keep",
		),
	})
	return nil
}
```
