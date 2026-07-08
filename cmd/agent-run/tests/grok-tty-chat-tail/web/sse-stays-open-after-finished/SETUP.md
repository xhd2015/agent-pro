# Scenario

**Bug**: W2 — SSE receives append after session status is already `finished`

```
seed finished session in AGENT_RUN_HOME
  -> open SSE with after=0
  -> sidecar appends CHAT_TAIL_SSE_AFTER_FINISHED_MARKER
  -> SSE must deliver appended row (WatchEvents ignores status)
```

## Steps

1. Start web (no grok POST needed).
2. Seed finished `grok-tty` session with one event.
3. Open SSE; sidecar appends marker after 500ms.

```go
import (
	"testing"
	"time"
)

const chatTailSSEAfterFinishedMarker = "CHAT_TAIL_SSE_AFTER_FINISHED_MARKER"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "sse-stays-open-after-finished"
	req.Mode = "sse-finished-append"
	req.SessionID = "chat_tail_sse_finished"
	startAgentRunWeb(t, req)
	seedFinishedSession(t, req, req.Runner, req.SessionID, "Finished session initial event")
	req.SSEAfterOffset = 0
	req.SSEMaxWait = 10 * time.Second
	runner := req.Runner
	sid := req.SessionID
	req.Sidecar = func() {
		time.Sleep(500 * time.Millisecond)
		appendSessionEvent(t, req, runner, sid, chatTailSSEAfterFinishedMarker)
	}
	return nil
}
```