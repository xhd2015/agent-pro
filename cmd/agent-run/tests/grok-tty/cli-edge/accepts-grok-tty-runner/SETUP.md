# Scenario

**Feature**: `grok-tty` is a recognized `--agent-runner` value

```
agent-run run --agent-runner grok-tty "hi" → not rejected as unknown runner; exits 0
```

## Preconditions

- Fake TUI responds to injected prompt with `Response: hi`.

## Steps

1. Set fake TUI to quick respond script.
2. Run `agent-run run --agent-runner grok-tty "hi"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokTTYCommand = fakeTUIRespondHi()
	req.Args = []string{"run", "--agent-runner", "grok-tty", "hi"}
	return nil
}
```