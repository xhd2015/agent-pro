# Scenario

**Feature**: S7 — `-e FOO=bar` → FOO=bar in Set

```
# S7
envEntries=["FOO=bar"]
  -> Set contains FOO=bar
  -> Unset empty
```

## Steps

1. Set EnvEntries to `FOO=bar`.
2. Assert FOO present with value bar.

## Context

- Color false: NO_COLOR not in Unset; FOO not stripped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvEntries = []string{"FOO=bar"}
	return nil
}
```
