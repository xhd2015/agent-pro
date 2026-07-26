# Scenario

**Bug**: second turn "what did I ask?" must produce assistant text that recalls first message "hi"

```
POST "hi" -> finished -> POST "what did I ask?" -> finished -> GET detail
```

## Preconditions

- `fake-codex` runner; history-aware prompt reaches the fake agent.

## Steps

1. Start web with `--token test`.
2. Create session with prompt `hi`; wait until `status=finished`.
3. POST follow-up `what did I ask?`; wait until `status=finished`.
4. `Run` GETs session detail for assertions.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "hi"
	req.FollowUpPrompt = "what did I ask?"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.SessionRunner, sessionID, "finished", 30*time.Second)

	postSessionMessage(t, req, req.SessionRunner, sessionID, req.FollowUpPrompt)
	waitForSessionStatus(t, req, req.SessionRunner, sessionID, "finished", 30*time.Second)

	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```