# Scenario

**Feature**: empty prompt is allowed with `--detach` (like `--open`)

```
agent-run run --agent-runner grok-tty --detach
  -> empty prompt OK; exit 0; both ids on stdout
```

## Preconditions

- With `--detach`, prompt optional (TTY only — non-TTY already rejected).

## Steps

1. Grouping documents prompt-policy class.
2. Leaf runs empty-prompt detach.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
