# Scenario

**Feature**: readonly capture sidecar accumulates fake TUI scrollback into events

```
fake TUI echoes Response: hi → stdout and sessions/codex-tty/.../events.jsonl contain hi
```

## Steps

1. Run `agent-run run --agent-runner codex-tty "hi"` with respond fake TUI.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexTTYCommand = fakeTUIRespondHi()
	req.Args = append(req.Args, "hi")
	return nil
}
```