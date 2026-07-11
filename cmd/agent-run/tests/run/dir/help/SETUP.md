# Scenario

**Feature**: `run --help` surfaces the `--dir` flag

```
agent-run run --help -> documents --dir; stdout ends with \n
```

## Steps

1. Leaves replace inherited run args with `run --help`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Help must not inherit --agent-runner fake-codex from run/ grouping.
	req.Args = []string{"run", "--help"}
	return nil
}
```
