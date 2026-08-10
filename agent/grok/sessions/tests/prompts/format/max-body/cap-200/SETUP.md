# Scenario

**Feature**: MaxBodyRunes=200 soft-caps a 220-x collapsed body

```
# 220 x runes; MaxBodySet MaxBodyRunes=200
Format -> ~200 content x's + "…"; not full 220
```

## Preconditions

- One user prompt of 220 `x` runes (no extra whitespace).
- MaxBodySet; MaxBodyRunes=200; no Grep.
- Op format-single.

## Steps

1. Write session with 220-x body.
2. Format with MaxBody 200.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idWhitespace
	req.MaxBodySet = true
	req.MaxBodyRunes = 200
	ts := atFixed(-10 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idWhitespace,
		Title:        "max-body 200",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt(longPromptRunes(220), ts)),
	})
	return nil
}
```
