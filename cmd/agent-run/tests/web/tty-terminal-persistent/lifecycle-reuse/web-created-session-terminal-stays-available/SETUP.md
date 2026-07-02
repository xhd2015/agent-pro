# Scenario

**Bug**: a real web-created codex-tty session loses its terminal after the first turn

```
POST /sessions runner=codex-tty -> generated web_* route + terminal_session_id session-N
finished session -> GET /terminal still available true
```

## Preconditions

- The web server is started by parent setup.
- A fake `codex` binary is placed on the web server PATH before session creation.
- The fake Codex TUI exits after answering, matching the reported finished chat state.

## Steps

1. Create a new codex-tty session through `POST /api/agent-run/sessions`.
2. Wait until session detail reports `status:"finished"` and a non-empty
   `terminal_session_id`.
3. Probe `/api/agent-run/sessions/codex-tty/<generated>/terminal`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	createWebCodexTTYSessionThroughAPI(t, req)
	req.Mode = "http"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```
