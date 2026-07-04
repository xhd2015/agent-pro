## Expected

- After Ctrl-C (`\x03`) on the watch PTY, the watch process exits promptly (no hang).
- The remote session remains registered and TCP-reachable (detach only, not kill).
- Exit code 0 on clean detach.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.TimedOut {
		t.Fatal("watch did not detach on Ctrl-C within 3s")
	}
	if !resp.RegistryExists {
		t.Fatalf("registry missing after watch detach, session %s", resp.SessionID)
	}
	if !resp.SessionRunning {
		t.Fatalf("remote session %s not reachable after watch Ctrl-C detach", resp.SessionID)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("watch exit code = %d after Ctrl-C detach, want 0", resp.ExitCode)
	}
}
```