## Expected

- Attach probe OK (WS handshake succeeds before timeout).
- No non-zero exit code from a hard failure (timeout is OK since attach blocks).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if !resp.AttachProbeOK && resp.ExitCode != 0 {
		t.Fatalf("attach failed: exit=%d stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}
```
