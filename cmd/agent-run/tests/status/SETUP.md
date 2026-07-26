# Scenario

**Subcommand**: `status` — runner and storage status summary

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Leaf `Setup` sets `req.Args` to `status`.
2. `Run` executes `agent-run status`.
3. `Assert` checks exit code 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"status"}
	return nil
}
```