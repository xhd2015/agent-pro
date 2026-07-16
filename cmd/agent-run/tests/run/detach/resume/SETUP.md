# Scenario

**Feature**: `agent-run resume --detach` reopens a keep-alive daemon without attach

```
seed bound+exited
  -> agent-run resume --detach <session-id> [followup?]
  -> exit 0; both ids; no attach; registry keep-alive
```

## Preconditions

- Session must be resume-ready (bound + exited) unless leaf seeds otherwise.
- Fake TUI hold so daemon outlives parent.

## Steps

1. Grouping documents resume-detach class.
2. Leaves seed exited meta and run resume --detach.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	return nil
}
```
