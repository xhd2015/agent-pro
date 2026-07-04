# Scenario

**Feature**: `bash -c 'echo yes'` prints output and exits promptly

```
harness PTY -> tty-watch run bash -c 'echo yes' -> prints yes -> exit 0 within 8s
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-bash-c-echo-exits"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```