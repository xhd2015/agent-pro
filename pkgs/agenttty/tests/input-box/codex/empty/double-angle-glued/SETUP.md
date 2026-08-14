# Scenario

**Feature**: Codex 0.146 » with footer glue on that line is empty

```
» Summarize recent commitsgpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> empty
```

## Preconditions

- Snapshot uses `»` (U+00BB) and must **not** contain legacy `›`.
- Same ` medium · ` glue rule as `›`.

## Steps

1. Inject a `»` + glued footer line.

## Context

Glyph variant only; occupancy rule is unchanged.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "» Summarize recent commitsgpt-5.6-terra medium · /private/var/folders/s_/nd3t_zbx61747w0qdryxh4wm0000gp…\n"
	req.Fixture = ""
	return nil
}
```
