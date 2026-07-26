# Scenario

**Feature**: `agent-run send` sends prompt via WebSocket same as `tty send`

```
# connect WS -> inject prompt with Ctrl+U prefix -> read response -> extract assistant text -> store event
agent-run send session-1 "hello" -> resolveTerminal -> WS connect -> inject Ctrl+U + "hello" + \r -> capture -> store events
```

## Steps

1. Start fake ptywrap WS server that records input and responds with scrollback.
2. Registry entry points to the fake server.
3. `req.Args` = `["send", "session-1", "hello"]`.
4. `req.Mode` = `"send-probe"` for extended timeout.
5. Assert that the send command succeeds (captures response).

```go
import (
	"fmt"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartFakePTYWrap = true
	req.FakePTYInputReceived = make(chan string, 10)
	startFakePTYWrapServer(t, req)
	waitForPortOpen(t, fmt.Sprintf("127.0.0.1:%d", req.FakePTYWrapPort), 5*time.Second)

	req.RegistryDir = "grok-tty-registry"
	req.RegistrySessionID = "session-1"
	req.RegistryEntryJSON = defaultRegistryEntryJSON(req.RegistrySessionID, req.FakePTYWrapPort)
	writeMockRegistryEntry(t, req)

	req.Args = []string{"send", req.RegistrySessionID, "hello"}
	req.Mode = "send-probe"
	req.ExecTimeout = 45 * time.Second
	return nil
}
```