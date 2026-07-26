# Scenario

**Feature**: session create stores server cwd on session meta as `workspace`

```
POST /api/agent-run/sessions -> GET detail -> session.workspace == web process cwd
```

## Preconditions

- Web server working directory is `req.TempDir`.
- Create uses `fake-codex` runner with a non-empty prompt.

## Steps

1. Start web server (`cmd.Dir = req.TempDir`).
2. POST create session.
3. `Run` GETs session detail.

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
	req.CreatePrompt = "workspace probe"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```