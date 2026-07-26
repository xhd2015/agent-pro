# Scenario

**Feature**: web follow-up returns accepted and message is delivered via queue

```
stub-tty idle -> POST messages -> queue drains -> text injected into PTY scrollback
```

## Steps

1. Start idle stub-tty; seed agent session mapping.
2. POST follow-up; poll registry snapshot/observer for delivered text.

```go
import (
	"encoding/json"
	"net/http"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	startAgentRunWeb(t, req)
	terminalID := startStubTTYBackground(t, req)
	req.TerminalSessionID = terminalID
	req.SessionID = "agent-session-deliver"
	seedAgentSessionForStubTTY(t, req, req.SessionID, terminalID)
	req.FollowUpPrompt = "deliver-via-queue"
	body, _ := json.Marshal(map[string]string{"message": req.FollowUpPrompt})
	req.HTTPMethod = http.MethodPost
	req.HTTPPath = "/api/agent-run/sessions/" + req.Runner + "/" + req.SessionID + "/messages"
	req.HTTPBody = string(body)
	req.Mode = "http"
	return nil
}
```
