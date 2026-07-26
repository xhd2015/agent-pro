# Scenario

**Feature**: help text documents session env injection flags

```
agent-run run --help    -> --prepend-path, -e/--env
agent-run resume --help -> --prepend-path, -e/--env
```

## Steps

1. Leaf sets `req.Args` to `run --help` or `resume --help`.
2. Assert stdout mentions the new flags and ends with trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaves finalize Args; default to run --help as a safe baseline.
	if len(req.Args) == 0 {
		req.Args = []string{"run", "--help"}
	}
	return nil
}
```
