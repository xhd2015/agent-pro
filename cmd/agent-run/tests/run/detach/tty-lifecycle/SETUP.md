# Scenario

**Feature**: `run --detach` on TTY: silent daemon start, soft bind, print both ids, keep-alive

```
agent-run run --agent-runner grok-tty --detach "prompt"
  -> silent start (no attach, no event stream)
  -> soft grok bind (miss OK)
  -> stdout: session-id + terminal-id
  -> registry kept for later attach/send
```

## Preconditions

- Default runner for this branch: `grok-tty` with fake TUI hold.
- Discovery soft-bind hit out of scope — silence of stream noise + soft miss are required.

## Steps

1. Grouping installs grok-tty fake TUI hold + common `--detach` args.
2. Leaves specialize asserts (silence / ids / registry / soft bind / status).

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "grok-tty"
	req.Prompt = "detach-lifecycle"
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
