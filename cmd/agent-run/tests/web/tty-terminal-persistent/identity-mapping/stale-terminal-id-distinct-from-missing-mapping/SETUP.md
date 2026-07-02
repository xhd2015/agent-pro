# Scenario

**Bug**: stale mapped PTY is not the same as absent terminal mapping

```
terminal_session_id session-1 -> stale registry listen_addr -> GET /terminal
  -> available false with terminal_session_id preserved
```

## Preconditions

- Metadata contains `terminal_session_id`.
- Registry file exists but points to a closed local port.

## Steps

1. Write mapped web chat metadata.
2. Write stale `codex-tty-registry/session-1.json`.
3. Fetch terminal status for the web chat id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, unusedLocalAddr(t))
	return nil
}
```
