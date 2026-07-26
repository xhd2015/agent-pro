# Scenario

**Feature**: `--detach` does not unlock `--no-submit` (still requires `--open`)

```
agent-run run --agent-runner grok-tty --detach --no-submit "x"
  -> exit ≠ 0
  -> error explains --no-submit requires --open
```

## Steps

1. Invoke run with `--detach --no-submit` on a TTY runner (no `--open`).

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
		"--no-submit",
		req.Prompt,
	}
	return nil
}
```
