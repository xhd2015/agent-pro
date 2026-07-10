# Scenario

**Feature**: invalid `--no-submit` combinations fail before starting a session

```
agent-run run --no-submit --agent-runner grok-tty "x" -> error (requires --open)
agent-run run --open --no-submit --agent-runner fake-codex "x" -> error (non-TTY)
```

## Preconditions

- Reject paths must fail validation (exit ≠ 0) without requiring a live TTY attach.

## Steps

1. Grouping documents reject mode.
2. Child dirs split missing `--open` vs non-TTY under open family.

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
