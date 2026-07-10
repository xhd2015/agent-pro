# Scenario

**Feature**: `run --help` documents `--open`

```
agent-run run --help -> usage lists --open
```

## Preconditions

- `agent-run` binary is built (root Setup).

## Steps

1. Grouping marks help mode (no runner required).
2. Leaf runs `run --help` and asserts flag text.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = ""
	req.Args = []string{"run", "--help"}
	return nil
}
```
