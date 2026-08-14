# Scenario

**Feature**: Codex » with a draft and footer on the next line is occupied

```
» EXP-DRAFT-NOTE-42
  gpt-5.6-terra medium · /private/…
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Uses `»` only (no `›`) so the 0.146 glyph path cannot hide behind `›`.

## Steps

1. Inject `»` + draft; footer on the following line.

## Context

Pair of `codex/empty/double-angle-glued` for the occupied `»` outcome.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "» EXP-DRAFT-NOTE-42\n  gpt-5.6-terra medium · /private/var/folders/s_/nd3t_zbx61747w0qdryxh4wm0000gp…\n"
	req.Fixture = ""
	return nil
}
```
