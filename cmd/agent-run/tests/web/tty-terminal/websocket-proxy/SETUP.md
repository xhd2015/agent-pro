# Scenario

**Feature**: terminal websocket proxy attaches browser clients to resolved ptywrap terminal

```
browser WS /api/agent-run/sessions/<runner>/<id>/terminal/ws
  -> agent-run auth + terminal resolver
  -> ptywrap WS /api/terminal
  -> PTY bytes, input, Enter, resize
```

## Preconditions

- Websocket attach endpoint is under authenticated `/api/agent-run` namespace.
- The client never supplies arbitrary upstream addresses.

## Steps

1. Start fake ptywrap websocket server.
2. Write matching tty session and registry fixture.
3. `Run` attaches through the agent-run web websocket endpoint.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "websocket"
	req.Runner = "codex-tty"
	req.SessionID = "tty-session-1"
	req.RegistryTranscript = "terminal-ready\n"
	req.WSAuth = req.WebToken
	req.WSInput = "hello from browser\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.WSPath = terminalWSPath(req.Runner, req.SessionID)
	return nil
}
```
