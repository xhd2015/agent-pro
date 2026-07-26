# Scenario

**Feature**: `run --detach --open` is rejected

```
agent-run run --agent-runner grok-tty --detach --open "x"
  -> exit ≠ 0
  -> stderr explains mutual exclusion
```

## Steps

1. Invoke run with both flags on `grok-tty` and a short prompt.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.Prompt = "x"
	setGrokTTYCommand(req, fakeTUIRespondHi())
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"--detach",
		"--open",
		req.Prompt,
	}
	return nil
}
```
