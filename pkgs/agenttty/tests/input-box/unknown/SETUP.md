# Scenario

**Feature**: snapshots with no live composer glyph classify as unknown

```
empty bytes | chrome without › » ❯
  -> DetectInputBox
  -> unknown
```

## Preconditions

- No Codex `›`/`»` and no Grok `❯` on any line.
- Distinct from empty-composer (glyph present, box unused).

## Steps

1. Mark `req.Family=unknown`.
2. Leaf injects empty text or glyph-free chrome.

## Context

Unknown is also the report token when TTY is unreachable / snapshot missing.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "unknown"
	return nil
}
```
