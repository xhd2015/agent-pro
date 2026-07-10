# Scenario

**Feature**: `--open` cannot be combined with `--json`

```
agent-run run --open --json --agent-runner grok-tty "x" -> exit ≠ 0
```

## Steps

1. Grouping marks json conflict class.
2. Leaf uses a TTY runner so failure is the flag conflict, not non-TTY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// JSON conflict class: leaf sets --open + --json on a TTY runner.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
