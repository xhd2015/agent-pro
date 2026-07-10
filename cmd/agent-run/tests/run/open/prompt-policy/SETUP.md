# Scenario

**Feature**: empty prompt is allowed only with `--open`

```
agent-run run --agent-runner <runner>           -> prompt is required
agent-run run --agent-runner grok-tty --open    -> empty prompt OK
```

## Preconditions

- Without `--open`, existing rule: prompt required after trim.
- With `--open`, prompt optional (TTY only — non-TTY already rejected).

## Steps

1. Grouping documents prompt-policy class.
2. Leaves set with/without `--open` and assert exit / error text.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Prompt-policy leaves choose with/without --open; ensure base is `run`.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
