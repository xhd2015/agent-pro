# Scenario

**Feature**: `status grok-tty/<session_id>` resolves seeded meta when unique

```
seed sessions/grok-tty/test-ref-s1/meta.json
  -> agent-run status grok-tty/test-ref-s1
  -> session line includes grok-tty/test-ref-s1
```

## Steps

1. Seed finished bound+exited meta.
2. Run status with `runner/session` ref.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "test-ref-s1"
	req.RunnerSessionID = "550e8400-e29b-41d4-a716-446655440222"
	req.MetaStatus = "finished"
	req.TerminalSessionID = "term-ref-1"
	req.InitialPrompt = "ref resolve"
	seedBoundExitedDeadTerminal(t, req)
	req.Args = []string{"status", req.Runner + "/" + req.SessionID}
	return nil
}
```
