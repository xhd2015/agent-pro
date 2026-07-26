# Scenario

**Feature**: `--open --no-submit` is rejected for non-TTY runners

```
agent-run run --open --no-submit --agent-runner fake-codex "x" -> exit ≠ 0
```

## Preconditions

- Runner is not a TTY backend (no adhoc PTY / registry attach path).
- Both `--open` and `--no-submit` are present so the open-family TTY gate applies.

## Steps

1. Grouping marks non-TTY reject class.
2. Leaf picks a concrete non-TTY runner and prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Non-TTY reject class: leaves pick runner; ensure we start from `run`.
	if len(req.Args) == 0 || req.Args[0] != "run" {
		req.Args = []string{"run"}
	}
	return nil
}
```
