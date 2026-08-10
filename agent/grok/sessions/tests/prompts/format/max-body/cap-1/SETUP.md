# Scenario

**Feature**: MaxBody N=1 yields one content rune + `…`

```
# long body "abcdef..."; MaxBodyRunes=1
Format -> "a…" (1 content rune + ellipsis outside N)
```

## Preconditions

- Body starts with distinct first rune `a` then more letters (length ≫ 1).
- MaxBodySet; MaxBodyRunes=1.
- Op format-single.

## Steps

1. Write session with multi-rune body.
2. Format with MaxBody 1.

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
	req.MaxBodyRunes = 1
	ts := atFixed(-10 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idWhitespace,
		Title:        "max-body 1",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt("abcdefghi", ts)),
	})
	return nil
}
```
