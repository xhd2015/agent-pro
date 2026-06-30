# Scenario

**Bug**: assistant timeline must use streaming phases during web runs

```
POST session -> run -> events.jsonl contains assistant message with phase start, update, end
```

## Preconditions

- Web create-session starts background `agentui.Run` with streaming emits enabled.

## Steps

1. Start web server; POST create with prompt `stream phases please`.
2. Poll `events.jsonl` until assistant phases `start`, `update`, and `end` all appear.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "stream phases please"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID

	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```