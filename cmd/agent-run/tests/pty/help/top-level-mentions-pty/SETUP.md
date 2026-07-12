# Scenario

**Feature**: top-level `agent-run --help` mentions the `pty` command

```
agent-run --help -> contains pty
```

## Steps

1. Leaf `Setup` sets `req.Args` to `--help`.
2. `Run` executes top-level help.
3. `Assert` checks exit 0 and stdout contains `pty`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
