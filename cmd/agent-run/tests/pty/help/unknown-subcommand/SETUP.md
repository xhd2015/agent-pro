# Scenario

**Feature**: unknown `pty` subcommand exits 1

```
agent-run pty not-a-real-pty-cmd -> exit 1
```

## Steps

1. Leaf `Setup` sets `req.Args` to an unknown pty subcommand.
2. `Run` executes the CLI.
3. `Assert` checks exit 1 and stderr indicates error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"pty", "not-a-real-pty-cmd"}
	return nil
}
```
