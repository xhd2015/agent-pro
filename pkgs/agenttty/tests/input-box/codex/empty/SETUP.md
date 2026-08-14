# Scenario

**Feature**: Codex last glyph line is an empty composer

```
last ›/» line contains " medium · "  OR  TrimSpace(remainder) == ""
  -> DetectInputBox
  -> empty
```

## Preconditions

- Occupied branch is the sibling: non-empty remainder **and** no footer glue.

## Steps

1. Mark family `codex-empty`.
2. Leaf supplies glued-placeholder, whitespace-only remainder, or `»` glue.

## Context

Live 0.147 empty is **not** a blank `›`; the hint is glued to the footer.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Family = "codex-empty"
	return nil
}
```
