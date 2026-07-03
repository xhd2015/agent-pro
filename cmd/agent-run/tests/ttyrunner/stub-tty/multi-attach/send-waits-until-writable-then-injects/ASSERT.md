## Expected

- `tty send` exits 0 after waiting for writable prompt.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.SendInjected { t.Fatalf("send failed exit=%d stderr=%s", resp.ExitCode, resp.Stderr) }
}
```
