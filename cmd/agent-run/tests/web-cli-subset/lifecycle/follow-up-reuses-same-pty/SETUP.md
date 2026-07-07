# Scenario

**Feature**: web follow-up reuses same PTY registry entry

```
POST codex-tty -> capture terminal_session_id -> POST follow-up -> registry ids unchanged
```

## Steps

1. Create codex-tty web session and capture terminal id.
2. POST follow-up message; list registry ids.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	createWebCodexTTYSession(t, req, "reuse same pty")
	waitForSessionStatus(t, req, req.Runner, req.ChatSessionID, "finished", 60*time.Second)
	req.TerminalSessionID = waitForTerminalSessionID(t, req, req.Runner, req.ChatSessionID, 10*time.Second)
	req.FollowUpPrompt = "follow-up same pty"
	postFollowUpMessage(t, req, req.Runner, req.ChatSessionID, req.FollowUpPrompt)
	req.Mode = "http"
	return nil
}
```
