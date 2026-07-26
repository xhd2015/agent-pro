# Scenario

**Bug**: follow-up sent to a live tty is marked finished immediately

```
finished web_* chat + mapped live PTY session-1
POST follow-up "what did I say?"
  -> backend writes prompt to existing PTY
  -> web session remains running until terminal response is observed
```

## Preconditions

- Session metadata stores `terminal_session_id:"session-1"`.
- The fake ptywrap accepts the prompt and stays connected.
- The test probes session metadata shortly after posting the follow-up.

## Steps

1. Start ptywrap-like websocket server.
2. Write finished mapped session metadata.
3. Write registry entry for the mapped PTY.
4. POST a follow-up message.
5. Fetch session detail and assert the follow-up is still running.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "followup"
	req.FollowUpPrompt = "what did I say?"
	req.RegistryTranscript = "followup-live-pty-ready\n"
	listenAddr := startControlFramePtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}
```
