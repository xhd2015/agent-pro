# Scenario

**Feature**: web POST messages on live TTY enqueues to agentsend queue

```
stub-tty background -> seed agent session meta -> POST messages -> send-queue jsonl
```

## Steps

1. Start stub-tty and seed agent session with terminal mapping.
2. POST follow-up message via web API.

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
	req.SessionID = "agent-session-send-queue"
	seedAgentSessionForStubTTY(t, req, req.SessionID, terminalID)
	req.FollowUpPrompt = "web-queue-probe"
	req.HTTPMethod = http.MethodPost
	req.HTTPPath = "/api/agent-run/sessions/" + req.Runner + "/" + req.SessionID + "/messages"
	body, _ := json.Marshal(map[string]string{"message": req.FollowUpPrompt})
	req.HTTPBody = string(body)
	req.Mode = "http"
	return nil
}
```
