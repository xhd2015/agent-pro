# Scenario

**Feature**: web terminal websocket write path reaches upstream PTY

```
browser WS binary input -> AttachRelay -> fake ptywrap records bytes
```

## Steps

1. Start web + fake ptywrap; seed session/registry.
2. Send keyboard bytes over terminal websocket.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	startAgentRunWeb(t, req)
	req.Runner = "codex-tty"
	req.SessionID = "web-tty-write"
	req.RegistryTranscript = "keyboard-test-ready"
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
