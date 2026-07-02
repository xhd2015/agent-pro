# Scenario

**Bug**: re-entering the chat route must attach to the same mapped PTY

```
open chat -> terminal websocket -> PTY session-1
navigate away/back -> terminal websocket -> same PTY session-1
```

## Preconditions

- Parent setup created one live mapped PTY.

## Steps

1. Attach websocket to `/terminal/ws` for the web chat id.
2. Fetch session detail to model leaving and returning to the chat route.
3. Attach websocket to `/terminal/ws` again.
4. Check both attaches reached the same fake PTY server.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ws"
	req.WSPath = terminalWSPath(req.Runner, req.ChatSessionID)
	return nil
}
```
