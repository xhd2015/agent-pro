# Scenario

**Feature**: after attached command exits, stdout is only the session id line

```
# attach then print id
harness PTY -> agent-term run true -> attach exits -> stdout: session-N only
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-pty"
	req.StartDaemon = true
	req.RunCommand = []string{"true"}
	return nil
}
```