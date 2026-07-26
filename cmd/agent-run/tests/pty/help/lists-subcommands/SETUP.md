# Scenario

**Feature**: `pty --help` lists all pty subcommands

```
agent-run pty --help -> stats, kill-orphans
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Leaf `Setup` sets `req.Args` to `pty --help`.
2. `Run` executes `agent-run pty --help`.
3. `Assert` checks exit 0, lists stats and kill-orphans, trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"pty", "--help"}
	return nil
}
```
