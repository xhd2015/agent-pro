# Scenario

**Feature**: occupancy uses the last composer glyph, not historical › above

```
› leftover draft
…scrollback…
› Summarize recent commits… medium · …
  -> DetectInputBox
  -> empty   (last glyph wins)
```

## Preconditions

- An earlier `› leftover` would be occupied if it were last.
- The live last line is the empty glued shape.

## Steps

1. Mark family `codex-last-glyph`.
2. Leaf injects historical draft plus a later empty glued line.

## Context

Historical composer lines in scrollback must not shadow the live box.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "codex-last-glyph"
	return nil
}
```
