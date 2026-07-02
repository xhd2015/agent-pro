# Scenario

**Bug**: status checks must resolve the stored terminal session id

```
meta: session_id web_persistent_terminal
meta: runner_session_id 019f...
meta: terminal_session_id session-1
codex-tty-registry/session-1.json live -> GET /terminal -> available true
```

## Preconditions

- No registry exists for the web chat id.
- No registry exists for the provider runner session id.

## Steps

1. Start a fake ptywrap server.
2. Write web chat metadata with distinct `runner_session_id` and `terminal_session_id`.
3. Write only `codex-tty-registry/session-1.json`.
4. Fetch terminal status for the web chat id.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RegistryTranscript = "mapped-terminal-ready\n"
	listenAddr := startMappedPtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}
```
