# Scenario

**Feature**: `--detach` and `--json` are mutually exclusive

```
run|resume --detach --json … -> exit ≠ 0
```

## Steps

1. Grouping marks json conflict class.
2. Leaves use TTY runner so failure is flag conflict, not non-TTY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// JSON conflict class: leaf sets --detach + --json on a TTY runner.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
