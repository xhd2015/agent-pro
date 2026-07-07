# Scenario

**Feature**: regression — web/stream SSE contract with CLI-parity event shape

```
POST fake-codex -> SSE after=0 -> user + assistant messages without phase field
```

## Steps

1. Mirror `web/stream/sse-delivers-new-events` with CLI-parity assertions.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "fake-codex"
	req.Prompt = "sse cli parity"
	sessionID, _, _ := postCreateSession(t, req, req.Runner, req.Prompt)
	req.SessionID = sessionID
	req.SSEAfterOffset = 0
	req.SSEMaxWait = 45 * time.Second
	req.Mode = "sse"
	return nil
}
```
