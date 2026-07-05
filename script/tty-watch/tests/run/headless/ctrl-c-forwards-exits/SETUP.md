# Scenario

**Feature**: headless Ctrl-C forwards SIGINT to PTY child and exits when child exits

```
# trap exits 0 on INT; harness SIGINTs headless parent (not PTY)
harness SIGINT -> tty-watch run --headless sh trap -> child exit 0 -> registry cleaned
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-headless-ctrl-c-forwards-exits"
	req.RunCommand = []string{"sh", "-c", "trap 'exit 0' INT; sleep 300"}
	return nil
}
```