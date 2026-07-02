# Scenario

**Bug**: terminal resolution must use terminal identity instead of chat or provider identity

```
web_* chat id + provider runner_session_id + terminal_session_id session-1
  -> GET /terminal -> resolve codex-tty-registry/session-1.json
```

## Preconditions

- `runner_session_id` is provider state and may not have a registry file.
- `terminal_session_id` is the only PTY registry key for this session.

## Steps

1. Descendant setup chooses live or stale registry behavior.
2. `Run` fetches terminal status for the web chat id.

## Context

- The tree intentionally does not create `codex-tty-registry/<web_chat_id>.json`.
- The tree intentionally does not create `codex-tty-registry/<runner_session_id>.json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "http"
	req.HTTPPath = terminalStatusPath(req.Runner, req.ChatSessionID)
	return nil
}
```
