# Scenario

**Feature**: WaitSession waits for remote PTY session exit via websocket

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartDaemon = true
	return nil
}
```