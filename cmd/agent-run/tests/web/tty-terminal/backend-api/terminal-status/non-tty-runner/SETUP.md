# Scenario

**Feature**: non-tty runners never advertise terminal attach

```
codex session -> GET /terminal -> available false
```

## Preconditions

- Only `codex-tty` and `grok-tty` use tty registry lookup.

## Steps

1. Write a non-tty `codex` session.
2. Fetch terminal status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex"
	req.SessionID = "non-tty-session"
	writeSessionFixture(t, req, req.Runner, req.SessionID, "finished")
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
