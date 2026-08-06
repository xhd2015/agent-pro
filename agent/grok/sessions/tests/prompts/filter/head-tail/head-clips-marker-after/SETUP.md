# Scenario

**Feature**: head 2 of 5 prints first two prompts and trailing omission marker

```
# p1..p5 chrono; Head=2
format-single -> [ts] p1 \n [ts] p2 \n (...3 omitted...)\n
OmittedAfter=3 OmittedBefore=0
```

## Preconditions

- Session with 5 chrono prompts p1..p5.
- HeadSet Head=2; Op format-single.
- ColorMode never (plain marker).

## Steps

1. Write 5 prompts.
2. Format with head 2.

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
	req.Head = 2
	req.ColorMode = "never"
	end := atFixed(-1 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterHead,
		Title:        "head clips",
		LastActiveAt: end,
		Updates: chronoPromptUpdates(end, "p1", "p2", "p3", "p4", "p5"),
	})
	return nil
}
```
