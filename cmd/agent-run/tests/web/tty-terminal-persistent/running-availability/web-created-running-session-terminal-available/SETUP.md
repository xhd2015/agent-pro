# Scenario

**Bug**: terminal status is unavailable while a web-created tty turn is still running

```
POST /sessions runner=codex-tty
  -> pty registry contains session-N
  -> GET /sessions/codex-tty/web_*/terminal before response completion
  -> available:true and terminal_session_id=session-N
```

## Preconditions

- The fake `codex` command sleeps before printing its final response.
- The probe happens after the PTY registry exists but before that final response.

## Steps

1. Create the running web session.
2. Wait for the backend PTY registry id.
3. Request terminal status for the web chat id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	createRunningWebCodexTTYSessionThroughAPI(t, req)
	waitForAnyRegistryID(t, req, 3_000_000_000)
	req.HTTPMethod = "GET"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```
