# Scenario

**Feature**: `bash -c 'echo yes'` leaves `[Terminal exited]` at column 0 and ends with newline

```
tty-watch run bash -c 'echo yes' -> PTY output:
  yes
  [Terminal exited]
(host prompt must not render far right)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-bash-c-exit-marker-column-zero"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```