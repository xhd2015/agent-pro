# Scenario

**Feature**: web follow-up on live TTY uses shared SendToAgentSession queue

```
POST /sessions/{runner}/{id}/messages -> ResolveByAgentSession -> agentsend.Enqueue
```

## Preconditions

- Background stub-tty session provides live writable terminal.
- Agent session metadata stores `terminal_session_id` mapping.

## Steps

1. Grouping setup sets `req.Area = "send"` and enables stub-tty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Area = "send"
	req.Runner = "stub-tty"
	req.Env = append(req.Env, "AGENT_RUN_ENABLE_STUB_TTY=1")
	return nil
}
```
