# Scenario

**Feature**: human list on empty store exits 0 with trailing newline

```
agent-run sessions (empty) -> optional header only; no session rows; exit 0
```

## Preconditions

- Fresh `AGENT_RUN_HOME` has no session directories.

## Steps

1. Run `agent-run sessions` with no extra flags (list mode default).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// list grouping already set Args to ["sessions"]; keep human mode (no --json).
	// Ensure home is empty of sessions (default after root Setup).
	req.Args = []string{"sessions"}
	return nil
}
```
