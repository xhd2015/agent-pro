# Scenario

**Feature**: tail 2 of 5 prints leading omission marker then last two prompts

```
# p1..p5; Tail=2
format-single -> (...3 omitted...)\n [ts] p4 \n [ts] p5\n
OmittedBefore=3 OmittedAfter=0
```

## Preconditions

- Session with 5 chrono prompts; TailSet Tail=2; Op format-single.

## Steps

1. Write 5 prompts.
2. Format with tail 2.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idFilterHead
	req.TailSet = true
	req.Tail = 2
	req.ColorMode = "never"
	end := atFixed(-1 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterHead,
		Title:        "tail clips",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end, "p1", "p2", "p3", "p4", "p5"),
	})
	return nil
}
```
