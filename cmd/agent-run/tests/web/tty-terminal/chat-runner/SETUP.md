# Scenario

**Feature**: existing chat sessions display runner as read-only metadata

```
route /sessions/<runner>/<session-id> -> session metadata runner
follow-up POST -> uses route runner, not mutable home runner select
```

## Preconditions

- Existing session runner is fixed session metadata.
- Chat page must not allow changing runner for follow-up messages.

## Steps

1. Seed an existing session.
2. Browser opens the session route.
3. UI verifies runner is visible but no session-page runner select is enabled.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "ui"
	req.Runner = "codex-tty"
	req.SessionID = "readonly-runner-session"
	writeSessionFixture(t, req, req.Runner, req.SessionID, "finished")
	return nil
}
```
