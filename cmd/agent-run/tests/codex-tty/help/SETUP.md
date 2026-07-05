# Scenario

**Subcommand**: `help` — top-level usage includes `attach` and `send` commands

```
agent-run --help → lists attach and send alongside web, run, sessions, status
```

## Preconditions

- `--help` exits 0 and prints usage to stdout.

## Steps

1. Leaf `Setup` sets `req.Args` to `--help`.
2. `Run` executes `agent-run --help`.
3. `Assert` checks stdout lists `attach`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	_ = t
	return nil
}
```