# Scenario

**Feature**: `bash -c 'echo yes'` on interactive TTY emits CRLF so lines start at column 0

```
tty-watch run bash -c 'echo yes' -> raw PTY bytes:
  yes\r\n[Terminal exited]\r\n
(without CR, host prompt renders far right after exit)
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-bash-c-exit-marker-column-zero"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```