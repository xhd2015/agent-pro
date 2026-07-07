# Scenario

**Feature**: SSE stream includes think and error events on bind failure

```
failure bind -> events.jsonl think + error
  -> SSE after=0 delivers both types to client
```

## Steps

1. Configure failure binding env and start web.
2. POST `grok-tty` session.
3. Collect SSE from offset 0 until stream ends.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Action = "sse-emits-think-and-error"
	configureBindingFailureEnv(t, req, "sse think error probe")
	startWebGrokSession(t, req)
	req.SSEAfterOffset = 0
	req.SSEMaxWait = bindingFailureFinishTimeout
	return nil
}
```