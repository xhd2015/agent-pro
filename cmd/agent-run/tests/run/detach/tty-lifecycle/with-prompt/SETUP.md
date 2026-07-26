# Scenario

**Feature**: `run --detach` with a prompt exits after registry (+ soft bind budget)

```
agent-run run --agent-runner grok-tty --detach "hi"
  -> exit 0; both ids on stdout
```

## Steps

1. Use grouping detach args with prompt `hi`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hi"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--detach", req.Prompt}
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	req.Mode = "detach-registry-after"
	return nil
}
```
