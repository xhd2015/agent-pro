# Scenario

**Feature**: invalid `--detach` combinations fail before starting a session

```
agent-run run --detach --open … -> error
agent-run run --detach --json … -> error
agent-run run --detach --agent-runner fake-codex … -> error
agent-run run --detach --no-submit … -> --no-submit requires --open
```

## Preconditions

- Reject paths must fail validation (exit ≠ 0) without requiring a live daemon.

## Steps

1. Grouping documents reject mode.
2. Child dirs split exclusivity / non-TTY / no-submit gate.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = []string{"run"}
	}
	return nil
}
```
