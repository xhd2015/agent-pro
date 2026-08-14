# Scenario

**Feature**: scrollback without › / » / ❯ is unknown

```
boot chrome + model line, no composer glyph
  -> DetectInputBox
  -> unknown
```

## Preconditions

- Text may mention `medium` or a path; it must not contain `›`, `»`, or `❯`.

## Steps

1. Inject glyph-free chrome as `req.Scrollback`.

## Context

Footer substring alone does not create a composer line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "Starting Codex\nmodel: gpt-5.6-terra\nworkspace /private/tmp/demo\n"
	req.Fixture = ""
	return nil
}
```
