# Scenario

**Feature**: last live Codex composer glyph is › or »

```
snapshot last › (U+203A) or » (U+00BB)
  -> DetectInputBox
  -> empty | occupied  (not unknown)
```

## Preconditions

- Last composer glyph in the snapshot is Codex, not Grok `❯`.
- Empty vs occupied is decided only on **that** glyph’s line.

## Steps

1. Mark `req.Family=codex` and `ProviderID=codex-tty`.
2. Child leaves inject glued-footer, draft, `»`, or historical glyphs.

## Context

Codex 0.146+ may render `»` instead of `›`. Same occupancy rule.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "codex"
	req.ProviderID = "codex-tty"
	return nil
}
```
