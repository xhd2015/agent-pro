# Scenario

**Feature**: `run --detach --json` is rejected

```
agent-run run --agent-runner grok-tty --detach --json "x"
  -> exit ≠ 0
  -> exclusivity error
```

## Steps

1. Invoke run with both flags on `grok-tty`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.Prompt = "x"
	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"--detach",
		"--json",
		req.Prompt,
	}
	return nil
}
```
