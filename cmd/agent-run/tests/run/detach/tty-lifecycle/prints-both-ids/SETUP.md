# Scenario

**Feature**: detach always prints both id lines on stdout (product contract C)

```
agent-run run --agent-runner grok-tty --detach "hi"
  -> stdout contains:
       session-id: <id>
       terminal-id: <id>
```

## Steps

1. Run detach with prompt; assert labeled lines on stdout (not stderr-only).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	return nil
}
```
