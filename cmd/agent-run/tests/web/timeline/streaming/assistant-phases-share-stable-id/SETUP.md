# Scenario

**Feature**: all phased assistant events in one turn reference the same `id`

```
single assistant turn -> multiple phase rows -> identical id field
```

## Preconditions

- Streaming emit assigns stable `AgentEvent.ID` per assistant reply.

## Steps

1. Same create-session flow as `streaming-message-phases-emitted`.
2. Assert collects distinct ids among phased assistant rows — exactly one id.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "stable stream id"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID

	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + sessionID
	return nil
}
```