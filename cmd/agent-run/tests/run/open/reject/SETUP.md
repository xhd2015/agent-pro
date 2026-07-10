# Scenario

**Feature**: invalid `--open` combinations fail before starting a session

```
agent-run run --open --agent-runner <non-tty> "x" -> error
agent-run run --open --json --agent-runner grok-tty "x" -> error
```

## Preconditions

- Reject paths must fail validation (exit ≠ 0) without requiring a live TTY attach.

## Steps

1. Grouping documents reject mode.
2. Child dirs split non-TTY vs `--open`+`--json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Leaves finalize flags; start from bare run.
	if len(req.Args) == 0 {
		req.Args = []string{"run"}
	}
	return nil
}
```
