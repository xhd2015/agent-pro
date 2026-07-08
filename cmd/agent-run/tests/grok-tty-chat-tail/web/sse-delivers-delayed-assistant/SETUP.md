# Scenario

**Bug**: W1 — web SSE delivers delayed assistant from grok-tty POST session

```
configure partial updates.jsonl + chrome hook
  -> POST grok-tty session
  -> SSE after=0 collects timeline
  -> scheduled completion append delivers CHAT_TAIL_ASSISTANT_MARKER
```

## Steps

1. Configure producer env (partial seed + chrome).
2. Start web; POST session with chat tail prompt.
3. Collect SSE until marker or timeout.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "sse-delivers-delayed-assistant"
	req.Mode = "sse"
	configureProducerKeepTTYEnv(t, req)
	startAgentRunWeb(t, req)
	req.SessionID = postCreateSession(t, req, req.Runner, req.Prompt)
	req.SSEAfterOffset = 0
	req.SSEMaxWait = sseFinishTimeout
	req.CompletionDelay = 3000 * time.Millisecond
	return nil
}
```