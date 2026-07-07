# Scenario

**Bug**: terminal status is unavailable while a web-created tty turn is still running

```
POST /sessions runner=grok-tty
  -> pty registry contains session-N
  -> GET /sessions/grok-tty/web_*/terminal before response completion
  -> available:true and terminal_session_id=session-N
```

## Preconditions

- Grok mock hook sleeps before printing its final response.
- The probe happens after the PTY registry exists but before that final response.

## Steps

1. Create the running web `grok-tty` session.
2. Wait for the backend PTY registry id.
3. Request terminal status for the web chat id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	createRunningWebGrokTTYSessionThroughAPI(t, req)
	waitForAnyRegistryID(t, req, 3_000_000_000)
	req.HTTPMethod = "GET"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```