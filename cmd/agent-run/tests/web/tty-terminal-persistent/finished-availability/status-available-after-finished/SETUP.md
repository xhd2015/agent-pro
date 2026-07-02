# Scenario

**Bug**: finished web tty session keeps terminal available while PTY is live

```
meta.status finished + terminal_session_id session-1
codex-tty-registry/session-1.json live -> GET /terminal -> available true
```

## Preconditions

- The session has already completed an agent turn.
- The backend PTY server is still alive.

## Steps

1. Start fake ptywrap server.
2. Write finished session metadata with terminal mapping.
3. Write live registry for `session-1`.
4. Fetch terminal status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RegistryTranscript = "finished-terminal-ready\n"
	listenAddr := startMappedPtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}
```
