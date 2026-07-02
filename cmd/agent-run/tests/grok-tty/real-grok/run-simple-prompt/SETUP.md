# Scenario

**Feature**: real grok run captures substantive assistant output

```
agent-run run --agent-runner grok-tty "say hi" → exit 0; stdout or events non-empty
```

## Steps

1. Run with prompt `say hi` against real grok TUI.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "grok-tty", "say hi"}
	return nil
}
```