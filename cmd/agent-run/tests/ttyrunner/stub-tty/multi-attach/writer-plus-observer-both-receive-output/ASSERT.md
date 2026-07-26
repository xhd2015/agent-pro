---
label: e2e
---

## Expected

- Both writer and observer receive PTY output.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	p := resp.MultiAttachProbe
	if len(p.WriterReceived) == 0 { t.Fatal("writer received no output") }
	if len(p.ObserverReceived) == 0 { t.Fatal("observer received no output") }
}
```
