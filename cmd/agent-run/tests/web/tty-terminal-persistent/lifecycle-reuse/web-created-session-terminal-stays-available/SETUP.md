# Scenario

**Bug**: a real web-created grok-tty session loses its terminal after the first turn

```
POST /sessions runner=grok-tty -> generated web_* route + terminal_session_id session-N
finished session -> GET /terminal still available true
```

## Preconditions

- Web started with grok mock harness by parent setup.
- Grok mock hook completes first turn and keeps PTY registry alive (`keep-tty`).

## Steps

1. Create a new `grok-tty` session through `POST /api/agent-run/sessions`.
2. Wait until session detail reports `status:"finished"` and a non-empty
   `terminal_session_id`.
3. Probe `/api/agent-run/sessions/grok-tty/<generated>/terminal`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	createWebGrokTTYSessionThroughAPI(t, req)
	req.Mode = "http"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```