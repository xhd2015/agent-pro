# Scenario

**Feature**: head clips *which* prompts; kept long body stays **full** by default

```
# two long prompts; Head=1; !MaxBodySet
format-single -> first full 220-x body + (...1 omitted...)
```

## Preconditions

- Session with 2 chrono long prompts (220 x's then 220 y's).
- HeadSet Head=1; no MaxBody.
- Op format-single; ColorMode never.

## Steps

1. Write two long prompts.
2. Format with head 1.

```go
import (
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idFilterHead
	req.HeadSet = true
	req.Head = 1
	req.ColorMode = "never"
	end := atFixed(-1 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterHead,
		Title:        "head long full",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end,
			longPromptRunes(220),
			strings.Repeat("y", 220),
		),
	})
	return nil
}
```
