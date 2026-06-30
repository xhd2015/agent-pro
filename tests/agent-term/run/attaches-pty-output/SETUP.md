# Scenario

**Bug**: `run` must attach TTY so remote command output is visible (not WaitSession poll only)

```
# interactive attach path
harness PTY -> agent-term run sh -> daemon PTY -> echo RUN_OK visible on harness PTY
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-pty"
	req.StartDaemon = true
	req.RunCommand = []string{"sh", "-c", "echo RUN_OK; exit 0"}
	return nil
}
```