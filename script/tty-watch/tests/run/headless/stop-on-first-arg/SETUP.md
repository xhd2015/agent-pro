# Scenario

**Feature**: `less-flags` StopOnFirstArg — command argv tokens are never parsed as flags

```
# --not-a-flag is passed to sh -c literally, not tty-watch
harness pipe -> tty-watch run --headless sh -c 'echo --not-a-flag'
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-headless-stop-on-first-arg"
	req.RunCommand = []string{"sh", "-c", "echo --not-a-flag"}
	return nil
}
```