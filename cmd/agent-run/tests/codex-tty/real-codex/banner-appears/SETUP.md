# Scenario

**Feature**: real codex TUI banner detected before prompt injection

```
agent-run run --agent-runner codex-tty "say hi" → no banner timeout; exit 0
```

## Steps

1. Run with real codex and short prompt (triggers banner wait + inject).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"run", "--agent-runner", "codex-tty", "say hi"}
	return nil
}
```