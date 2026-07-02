# Scenario

**Feature**: terminal status resolves only tty-backed sessions with live registries

```
sessions/<runner>/<session-id>/meta.json + <runner>-registry/<session-id>.json
  -> GET /terminal -> available true only when runner is tty and registry is reachable
```

## Preconditions

- TTY runners use `codex-tty` or `grok-tty`.
- Non-tty runners do not use tty registry discovery.

## Steps

1. Leaf setup writes session metadata.
2. Leaf setup optionally writes a live, missing, stale, or non-tty registry case.
3. `Run` fetches terminal status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Runner == "" {
		req.Runner = "codex-tty"
	}
	if req.SessionID == "" {
		req.SessionID = "tty-session-1"
	}
	req.HTTPPath = terminalStatusPath(req.Runner, req.SessionID)
	return nil
}
```
