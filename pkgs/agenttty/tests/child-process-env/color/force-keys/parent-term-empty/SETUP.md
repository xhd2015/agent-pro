# Scenario

**Feature**: S2 — color with empty parentTERM → TERM=xterm-256color

```
# S2 empty rewrite
parentTERM="", color=true
  -> TERM=xterm-256color in Set
```

## Steps

1. Set ParentTERM to empty string.
2. Assert force keys and TERM default.

## Context

- Empty effective TERM is treated like dumb for rewrite.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ParentTERM = ""
	return nil
}
```
