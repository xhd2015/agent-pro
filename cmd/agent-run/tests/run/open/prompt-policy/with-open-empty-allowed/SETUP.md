# Scenario

**Feature**: `--open` without a prompt is allowed on a TTY runner

```
agent-run run --agent-runner grok-tty --open
  -> must NOT fail with "prompt is required"
  -> (with AGENT_RUN_OPEN_ATTACH_INSTANT=1) completes open lifecycle
```

## Preconditions

- Fake TUI installed; open attach instant hook so the process does not block on
  interactive attach.

## Steps

1. Configure grok-tty fake TUI + instant attach.
2. Run `run --agent-runner grok-tty --open` with no prompt args.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Runner = "grok-tty"
	req.OpenInstantAttach = true
	setGrokTTYCommand(req, fakeTUIHoldSeconds(5))
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open"}
	req.ExecTimeout = 45 * time.Second
	return nil
}
```
