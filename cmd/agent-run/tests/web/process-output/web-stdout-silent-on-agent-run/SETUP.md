# Scenario

**Bug**: background web agent run must not leak formatted agent output to the web server process streams

```
POST fake-codex session -> wait finished -> captured web stdout/stderr lack 💬 and [done]
```

## Preconditions

- Open API (`WebTokenMode=omit`) matches typical local dev (no Bearer).
- Web process stdout/stderr buffers accumulate for the full test.

## Steps

1. Start `agent-run web --port 0` without `--token`.
2. POST create session (`fake-codex`, prompt `hi`).
3. Wait until session `status=finished`.
4. `Run` issues a no-op health GET so the leaf has a response.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "omit"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "hi"
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.SessionRunner, sessionID, "finished", 30*time.Second)

	req.HTTPMethod = "GET"
	req.HTTPAuth = ""
	req.HTTPPath = "/api/agent-run/health"
	return nil
}
```