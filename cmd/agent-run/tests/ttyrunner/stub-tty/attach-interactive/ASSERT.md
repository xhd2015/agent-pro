---
label: e2e
---

## Expected

- First interactive attach accepts keyboard input.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe == nil || !resp.MultiAttachProbe.WriterCanWrite {
		t.Fatal("expected interactive attach to accept input")
	}
}
```
