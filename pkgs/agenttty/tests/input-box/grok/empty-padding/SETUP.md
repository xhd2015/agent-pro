# Scenario

**Feature**: Grok ❯ with padding-only remainder is empty

```
│ ❯        │
  -> DetectInputBox
  -> empty
```

## Preconditions

- Last `❯` line has only whitespace after the glyph (before optional box border
  is not required; this leaf uses glyph + spaces).

## Steps

1. Inject `❯` plus trailing spaces.

## Context

Conservative Grok: padding-only after `❯` is empty. No footer-glue shortcut.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "Turn completed in 3.0s.\n❯     \n Shift+Tab:mode\n"
	req.Fixture = ""
	return nil
}
```
