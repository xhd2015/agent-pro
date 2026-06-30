# Scenario

**Subcommand**: `help` — usage text for top-level and subcommand help

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- `--help` exits 0 and prints usage to stdout.

## Steps

1. Leaf `Setup` sets `req.Args` to the help invocation.
2. `Run` executes `agent-run` with those args.
3. `Assert` checks stdout contains expected command names and flags.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = t
	return nil
}
```