# Scenario

**Feature**: `status <session_id>` resolves seeded meta (flat layout; bare id)

```
seed sessions/test-ref-s1/meta.json (runner=grok-tty)
  -> agent-run status test-ref-s1
  -> session line includes grok-tty/test-ref-s1 (display) or bare id
```

## Steps

1. Seed finished bound+exited meta under flat `sessions/<id>/`.
2. Run status with bare session id (compound `runner/id` refs are rejected).

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
	req.Args = []string{"status", req.SessionID}
	return nil
}
```
