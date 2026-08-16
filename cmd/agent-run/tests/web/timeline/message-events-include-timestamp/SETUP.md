# Scenario

**Feature**: initial user prompt persisted with millisecond `timestamp` on session detail API

```
POST /api/agent-run/sessions {prompt} -> appendUserPromptEvent(timestamp) -> GET detail -> events[]
```

## Preconditions

- Web server with explicit Bearer token.
- `fake-codex` runner on `PATH`.

## Steps

1. Start `agent-run web --token test --port 0`.
2. POST create session with prompt `timestamp check user`.
3. `Run` GETs session detail (leaf assert polls until user message has `timestamp` > 0).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "timestamp check user"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + sessionID
	return nil
}
```