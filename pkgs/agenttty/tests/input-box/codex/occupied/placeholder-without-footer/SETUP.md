# Scenario

**Feature**: placeholder-looking text without footer glue is occupied

```
› Summarize recent commits
  -> DetectInputBox
  -> occupied
```

## Preconditions

- Remainder after `›` is the hint string only — **no** ` medium · ` on that line
  or later.

## Steps

1. Inject `› Summarize recent commits` as `req.Scrollback`.

## Context

The empty rule needs the footer glue (or TrimSpace-empty remainder). Hint text
alone cannot be distinguished from a user draft.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Scrollback = "› Summarize recent commits\n"
	req.Fixture = ""
	return nil
}
```
