# Scenario

**Feature**: WaitSession blocks until a short-lived command exits

Unit-level reproduction of the websocket read-timeout retry bug in
`ptywrap/client.WaitSession` (used by `agent-term run`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "wait-session"
	req.StartDaemon = true
	req.RunCommand = []string{"sleep", "2"}
	return nil
}
```