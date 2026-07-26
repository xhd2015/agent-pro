---
label: e2e
---

## Expected

- `tty send` exits 0 after waiting for writable prompt.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.SendInjected { t.Fatalf("send failed exit=%d stderr=%s", resp.ExitCode, resp.Stderr) }
}
```
