# Scenario

**Feature**: arbitrary command output captured via WebSocket

```
# echo through PTY
REST create echo -> WS attach -> binary output contains hello
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "spawn-cmd"
	req.Command = []string{"echo", "hello"}
	req.Name = "echo-test"
	return nil
}
```