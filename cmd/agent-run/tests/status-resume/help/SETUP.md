# Scenario

**Feature**: help surfaces for top-level, `status`, and `resume`

```
agent-run --help -> lists resume
agent-run status --help -> session-ref / layers
agent-run resume --help -> --open, session-id
```

## Steps

1. Leaf sets `req.Args` to the help invocation.
2. `Run` executes CLI; assert documents new command/flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Leaves finalize Args; default to top-level help as a safe baseline.
	if len(req.Args) == 0 {
		req.Args = []string{"--help"}
	}
	return nil
}
```
