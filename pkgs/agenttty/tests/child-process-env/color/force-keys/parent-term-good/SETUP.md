# Scenario

**Feature**: S3 — color with good parentTERM preserves TERM (not dumb path)

```
# S3 preserve good TERM
parentTERM=xterm, color=true
  -> force keys + TERM=xterm (not forced to xterm-256color via dumb path)
```

## Steps

1. Set ParentTERM to `xterm` (usable, non-empty, non-dumb).
2. Assert force keys; TERM is `xterm` (or at least not the empty/dumb rewrite to a different value incorrectly).

## Context

- Old ApplyChildProcessEnv still emits `TERM=<effective>` when color and good TERM;
  policy must not rewrite good → xterm-256color only when empty/dumb.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ParentTERM = "xterm"
	return nil
}
```
