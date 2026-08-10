# Scenario

**Feature**: grep + MaxBody windows body around first match within N runes

```
# 150 a's + MATCH + 150 b's; Grep=MATCH; MaxBodyRunes=40
format-single -> snippet ≤ ~40 content runes around MATCH; side … if cut
```

## Preconditions

- Same long middle-match body as full-body leaf.
- GrepSet Grep=`MATCH`; MaxBodySet MaxBodyRunes=40; ColorMode never (plain).
- Op format-single.

## Steps

1. Write long prompt with MATCH in the middle.
2. Format with grep + MaxBody 40.

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
	req.Grep = "MATCH"
	req.MaxBodySet = true
	req.MaxBodyRunes = 40
	req.ColorMode = "never"
	raw := strings.Repeat("a", 150) + "MATCH" + strings.Repeat("b", 150)
	ts := atFixed(-5 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFilterSingle,
		Title:        "grep max-body window",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt(raw, ts)),
	})
	return nil
}
```
