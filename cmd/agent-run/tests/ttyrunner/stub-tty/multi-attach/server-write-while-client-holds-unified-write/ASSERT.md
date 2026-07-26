---
label: e2e
---

## Expected

- Server `tty send` succeeds concurrently with client unified writer.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.ServerSendWhileWriter {
		t.Fatal("server send should succeed while writer holds unified write")
	}
}
```
