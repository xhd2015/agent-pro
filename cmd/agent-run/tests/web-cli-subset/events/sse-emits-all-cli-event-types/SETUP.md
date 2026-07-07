# Scenario

**Feature**: SSE emits same event types CLI writes (not only ActionMessage)

```
POST fake-codex -> SSE after=0 -> includes terminal done event type
```

## Steps

1. Create fake-codex session via web.
2. Collect SSE events from offset 0 until stream ends.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "fake-codex"
	req.Prompt = "emit done event"
	sessionID, _, _ := postCreateSession(t, req, req.Runner, req.Prompt)
	req.SessionID = sessionID
	req.SSEAfterOffset = 0
	req.SSEMaxWait = 45 * time.Second
	req.Mode = "sse"
	return nil
}
```
