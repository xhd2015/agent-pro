# Scenario

**Feature**: `--detach` without a prompt is allowed on a TTY runner

```
agent-run run --agent-runner grok-tty --detach
  -> must NOT fail with "prompt is required"
  -> exit 0; stdout session-id: + terminal-id:
```

## Preconditions

- Fake TUI hold so keep-alive daemon outlives parent exit.

## Steps

1. Configure grok-tty fake TUI hold.
2. Run `run --agent-runner grok-tty --detach` with no prompt args.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach"}
	req.Mode = "detach-registry-after"
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
