# Scenario

**Feature**: color policy forces TTY color env and may rewrite TERM

```
# color true
BuildChildProcessEnv(…, color=true, parentTERM=…)
  -> Unset: NO_COLOR
  -> Set: FORCE_COLOR=1, CLICOLOR=1, CLICOLOR_FORCE=1, TERM=…
```

## Steps

1. Grouping sets Color=true for all color/* leaves.
2. Sub-groups choose TERM / user-NO_COLOR variants.

## Context

- Color policy applied last (wins over parent + user `-e` for force keys).
- User `-e NO_COLOR=…` is dropped from Set so Unset can clear parent NO_COLOR too.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Mode = "build"
	req.Color = true
	if req.RunnerID == "" {
		req.RunnerID = "grok-tty"
	}
	return nil
}
```
