# Scenario

**Feature**: S2 — color with parentTERM=dumb → TERM=xterm-256color

```
# S2 dumb rewrite
parentTERM=dumb, color=true
  -> BuildChildProcessEnv
  -> Unset NO_COLOR; Set FORCE_* + TERM=xterm-256color
```

## Steps

1. Set ParentTERM to `dumb`.
2. Assert force keys and TERM rewrite.

## Context

- Matches old ApplyChildProcessEnv color path for dumb parent TERM.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ParentTERM = "dumb"
	return nil
}
```
