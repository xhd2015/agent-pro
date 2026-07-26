# Scenario

**Feature**: after detach, session status remains `running` and terminal is live

```
agent-run run --agent-runner grok-tty --detach "hi"
  -> exit 0
  -> sessions/<id>/meta.json status=running
  -> registry terminal reachable
```

## Steps

1. Run detach with hold TUI.
2. Read meta + registry; assert status running and TCP live.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hi"
	req.Mode = "read-meta+registry"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	setGrokTTYCommand(req, fakeTUIHoldSeconds(45))
	return nil
}
```
