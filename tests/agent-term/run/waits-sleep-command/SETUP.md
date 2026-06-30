# Scenario

**Feature**: run waits for long-running commands without websocket read panic

Reproduces `agent-term run grok` panic: `repeated read on failed websocket connection`.
Uses `sleep 2` as a portable stand-in for any command that outlives the first WS
read deadline (2s in `WaitSession`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-exit-id"
	req.StartDaemon = true
	req.RunCommand = []string{"sleep", "2"}
	return nil
}
```