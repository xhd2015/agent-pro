# Scenario

**Feature**: grep + long body without MaxBody prints **full** collapsed body and highlights match

```
# 150 a's + MATCH + 150 b's; Grep=MATCH; ColorMode=always; !MaxBodySet
format-single -> full ~305-rune body; bold-red MATCH; no body window ellipsis
```

## Preconditions

- One prompt longer than 200 runes with match in the middle.
- GrepSet Grep=`MATCH`; ColorMode=`always`; no MaxBody.
- Op format-single (filter keeps the line; format shows full body + highlight).

## Steps

1. Write long prompt containing MATCH.
2. Format with grep + color, full body default.

```go
import (
	"strings"
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
	// Longer than old 200 soft-window so full-body default is observable.
	raw := strings.Repeat("a", 150) + "MATCH" + strings.Repeat("b", 150)
	ts := atFixed(-5 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep long full",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt(raw, ts)),
	})
	return nil
}
```
