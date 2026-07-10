# Scenario

**Feature**: `--open` is rejected for non-TTY runners

```
agent-run run --open --agent-runner fake-codex "x" -> exit ≠ 0
```

## Preconditions

- Runner is not a TTY backend (no adhoc PTY / registry attach path).

## Steps

1. Grouping marks non-TTY reject class.
2. Leaf picks a concrete non-TTY runner and prompt.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Non-TTY reject class: leaves pick runner; ensure we start from `run`.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
