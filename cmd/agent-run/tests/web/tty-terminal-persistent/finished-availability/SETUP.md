# Scenario

**Bug**: terminal availability is independent of finished chat status

```
finished web chat + terminal_session_id session-1 + live registry
  -> GET /terminal -> available true
```

## Preconditions

- Chat `status` is `finished`.
- The mapped PTY registry entry is still reachable.

## Steps

1. Descendant setup writes finished metadata and live registry.
2. `Run` fetches terminal status.

## Context

- This covers the missing terminal icon root cause where UI/API coupled terminal
  affordance to active chat turn status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "http"
	req.Status = "finished"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```
