# Scenario

**Feature**: head N ≥ total prints all lines and no omission marker

```
# 3 prompts; Head=10
format-single -> all 3 lines; no "(... omitted...)"
OmittedAfter=0
```

## Preconditions

- 3 prompts; Head=10; Op format-single.

## Steps

1. Write 3 prompts.
2. Format with head 10.

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
	req.Head = 10
	req.ColorMode = "never"
	end := atFixed(-1 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterHead,
		Title:        "head covers all",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end, "a", "b", "c"),
	})
	return nil
}
```
