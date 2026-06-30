# Scenario

**Feature**: SSE subscription receives user and assistant events for a new web session

```
POST create -> SSE after=0 -> user message + assistant message events
```

## Preconditions

- Create-session appends user event then runs agent (assistant events follow).

## Steps

1. Start web; POST create session with `sse hello`.
2. `Run` collects SSE events from offset 0 until connection ends or timeout.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.CreatePrompt = "sse hello"
	req.SSEAfterOffset = 0
	startWebServer(t, req)

	sessionID, _, _ := postCreateSession(t, req, req.SessionRunner, req.CreatePrompt)
	req.SessionID = sessionID
	req.SSEMaxWait = 45 * time.Second
	return nil
}
```