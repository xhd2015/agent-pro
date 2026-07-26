# Scenario

**Bug**: creating a web session must persist the initial user prompt in `events`

```
POST /api/agent-run/sessions {prompt: fix the bug} -> user message event -> GET detail
```

## Preconditions

- `fake-codex` available on `PATH` (root Setup builds it).
- Web server started with `--token test`.

## Steps

1. Start `agent-run web --token test --port 0`.
2. POST create session with runner `fake-codex` and prompt `fix the bug`.
3. `Run` GETs session detail for the created id.

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
	req.CreatePrompt = "fix the bug"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```