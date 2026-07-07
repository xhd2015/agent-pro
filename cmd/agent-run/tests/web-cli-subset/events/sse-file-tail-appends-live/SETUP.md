# Scenario

**Feature**: SSE tails events.jsonl and delivers rows appended after subscribe

```
POST running session -> SSE after=0 -> sidecar appends event -> SSE receives it
```

## Steps

1. Seed running session with one event; start web.
2. Open SSE; sidecar appends second event while connection open.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "fake-codex"
	req.SessionID = "sse-tail-live"
	seedRunningSessionForPrint(t, req, req.Runner, req.SessionID)
	req.SSEAfterOffset = 0
	req.SSEMaxWait = 8 * time.Second
	go func() {
		time.Sleep(400 * time.Millisecond)
		appendEventWhileSSE(t, req, req.Runner, req.SessionID, "Tail appended while SSE open")
	}()
	req.Mode = "sse"
	return nil
}
```
