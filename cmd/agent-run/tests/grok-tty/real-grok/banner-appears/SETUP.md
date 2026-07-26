# Scenario

**Feature**: real grok TUI banner detected before prompt injection

```
agent-run run --agent-runner grok-tty "say hi" → no banner timeout; exit 0
```

## Steps

1. Run with real grok and short prompt (triggers banner wait + inject).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--agent-runner", "grok-tty", "say hi"}
	return nil
}
```