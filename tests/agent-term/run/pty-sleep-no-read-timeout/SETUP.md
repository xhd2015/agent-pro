# Scenario

**Bug**: PTY `agent-term run sleep N` errors with `read tcp ... i/o timeout` after ~2s idle

Reproduces `agent-term run grok` failing when the remote process goes quiet.
Root cause: `readSessionID` leaves a 2s websocket read deadline active on the
connection; `readTerminalOutput` then times out during silence.

```
# interactive attach with quiet long-running command
harness PTY -> agent-term run sleep 3 -> attach bridge -> sleep exits -> session id
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-pty"
	req.StartDaemon = true
	req.RunCommand = []string{"sleep", "3"}
	return nil
}
```