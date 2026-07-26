# Scenario

**Feature**: list fails when daemon is not running

```
# no server
agent-term list -> connection error -> mentions agent-term serve
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "list-no-daemon"
	req.Args = []string{"list"}
	req.EnsureNoDaemon = true
	port, _ := pickFreePort(47681)
	req.DaemonPort = port
	return nil
}
```