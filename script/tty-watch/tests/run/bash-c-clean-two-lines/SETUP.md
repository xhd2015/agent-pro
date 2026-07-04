# Scenario

**Feature**: `bash -c 'echo yes'` shows exactly two visible lines without smear

```
tty-watch run bash -c 'echo yes' -> visible lines:
  yes
  [Terminal exited]
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-bash-c-clean-output"
	req.RunCommand = []string{"bash", "-c", "echo yes"}
	return nil
}
```