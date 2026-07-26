# Scenario

**Feature**: tty session without registry is gracefully unavailable

```
codex-tty session + no codex-tty-registry/<id>.json -> GET /terminal -> available false
```

## Preconditions

- Missing registry must not be treated as a server error.

## Steps

1. Write tty session metadata only.
2. Fetch terminal status.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "missing-terminal-session"
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
