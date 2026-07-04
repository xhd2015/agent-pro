# Scenario

**Feature**: Ctrl-C (`\x03`) forwards interrupt to PTY child

```
# trap handler prints marker when child receives INT
harness PTY -> tty-watch sh trap -> write \x03 -> TTY_WATCH_INTERRUPTED
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-ctrl-c"
	req.SendCtrlC = true
	req.RunCommand = []string{"sh", "-c", `trap 'echo TTY_WATCH_INTERRUPTED' INT; sleep 300`}
	return nil
}
```