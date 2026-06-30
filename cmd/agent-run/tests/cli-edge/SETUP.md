# Scenario

**Subcommand**: `cli-edge` — invalid subcommands and unknown agent runners

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Tests expect non-zero exit codes and actionable stderr messages.

## Steps

1. Leaf `Setup` sets `req.Args` for the invalid invocation.
2. `Run` executes `agent-run` with those args.
3. `Assert` checks exit code 1 and stderr content.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = t
	return nil
}
```