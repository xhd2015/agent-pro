# Scenario

**Feature**: ColorMode always + grep highlights the match span with ANSI

```
# prompt "hello MATCH world"; Grep=MATCH; ColorMode=always
format-single -> body contains CSI sequences around match
```

## Preconditions

- One matching prompt; GrepSet Grep=`MATCH`; ColorMode=`always`.
- Op format-single.
- Without color, match still shown; with always, `\x1b[` present near match.

## Steps

1. Write session with one prompt containing MATCH.
2. Format with grep + color always.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idFilterSingle
	req.GrepSet = true
	req.Grep = []string{"MATCH"}
	req.ColorMode = "always"
	ts := atFixed(-5 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "color grep",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt("hello MATCH world", ts)),
	})
	return nil
}
```
