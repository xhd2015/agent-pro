# Scenario

**Subcommand**: `help` — usage text for top-level and subcommand help

## Preconditions

- Top-level help leaf uses L2 `req.Mode = "handle"` (`agentruncli.Handle`); no binary.
- `--help` succeeds (exit 0) and prints usage to stdout.

## Steps

1. Leaf `Setup` sets `Mode: "handle"` and `req.Args` for the help invocation.
2. `Run` calls Handle in-process (short-path) or execs binary when Mode is empty.
3. `Assert` checks stdout contains expected command names and flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = t
	return nil
}
```
