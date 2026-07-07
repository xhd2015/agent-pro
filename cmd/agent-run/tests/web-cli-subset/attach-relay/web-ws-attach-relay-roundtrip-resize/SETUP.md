# Scenario

**Feature**: web terminal websocket uses AttachRelay with attach mode (IO + resize)

```
fake ptywrap (attach only) -> web terminal/ws -> transcript + resize JSON
```

## Steps

1. Start web server and fake ptywrap requiring `attach_mode=attach`.
2. Seed tty session + registry; attach websocket with resize + input.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "codex-tty"
	req.SessionID = "web-tty-1"
	req.RegistryTranscript = "terminal-ready-from-pty"
	req.WSAttachMode = "attach"
	addr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running", req.SessionID)
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, addr)
	req.WSPath = terminalWSPath(req.Runner, req.SessionID)
	req.WSResizeJSON = `{"type":"resize","cols":100,"rows":30}`
	req.WSInput = "resize-probe\n"
	req.Mode = "websocket"
	return nil
}
```
