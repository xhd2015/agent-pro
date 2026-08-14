# Scenario

**Feature**: leftover › above a settled empty glued line is empty

```
› leftover draft still in scrollback
assistant output
› Summarize recent commitsgpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> empty
```

## Preconditions

- First `› leftover draft` would classify occupied in isolation.
- Last `›` line is the live 0.147 empty glue shape.

## Steps

1. Inject both composer lines as `req.Scrollback`.

## Context

First-glyph-wins would wrongly report occupied.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "› leftover draft still in scrollback\nassistant output here\n› Summarize recent commitsgpt-5.6-terra medium · /private/var/folders/s_/nd3t_zbx61747w0qdryxh4wm0000gp…\n"
	req.Fixture = ""
	return nil
}
```
