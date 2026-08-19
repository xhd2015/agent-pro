# Scenario

**Feature**: Grok boxed empty composer `│ ❯ … │` is empty

```
 │ ❯                                                        │
  -> DetectInputBox
  -> empty
```

Live Grok pads the composer inside a box; the right border is not draft.

## Steps

1. Inject boxed empty `❯` line (border after glyph).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "pong\n    Worked for 6.0s\n │ ❯                                                        │\n Shift+Tab:mode\n"
	req.Fixture = ""
	return nil
}
```
