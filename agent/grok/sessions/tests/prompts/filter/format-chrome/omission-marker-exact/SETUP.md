# Scenario

**Feature**: omission line is the exact substring `(...M omitted...)`

```
# 4 prompts head 1 -> marker (...3 omitted...) exactly
```

## Preconditions

- 4 prompts; Head=1; format-single.
- Assert exact marker string form (parentheses, three dots each side, space around omitted).

## Steps

1. Write 4 prompts.
2. Format with head 1.

```go
import (
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
		Title:        "marker exact",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end, "only", "x2", "x3", "x4"),
	})
	return nil
}
```
