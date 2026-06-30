# Scenario

**Feature**: assistant timeline messages expose `timestamp` after background agent run

```
POST session -> agentui.Run emit(message, timestamp) -> GET detail -> assistant message timestamp > 0
```

## Preconditions

- `fake-codex` completes and appends assistant `message` events.
- Session `status` becomes `finished` when the run ends.

## Steps

1. Start web server, POST create session with prompt `hi`.
2. Wait until session status is `finished` and assistant message has `timestamp` > 0.
3. `Run` performs final GET for assert.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "hi"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.SessionRunner, sessionID, "finished", 30*time.Second)
	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```