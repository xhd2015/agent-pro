# Scenario

**Feature**: regression — web/tty-terminal websocket-proxy round-trip still works with attach mode

```
fake ptywrap -> web terminal/ws -> transcript + browser input echo
```

## Steps

1. Mirror `web/tty-terminal/websocket-proxy/round-trip-io` with attach relay backend.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "codex-tty"
	req.SessionID = "regression-ws-1"
	req.RegistryTranscript = "terminal-ready-from-pty"
	req.WSAttachMode = "attach"
	addr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running", req.SessionID)
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, addr)
	req.WSPath = terminalWSPath(req.Runner, req.SessionID)
	req.WSInput = "hello from browser\n"
	req.Mode = "websocket"
	return nil
}
```
