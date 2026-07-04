# Scenario

**Feature**: short-lived `echo` prints output and exits promptly

```
harness PTY -> tty-watch run echo yes -> prints yes -> exit 0 within 8s
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-echo-exits"
	req.RunCommand = []string{"echo", "yes"}
	return nil
}
```