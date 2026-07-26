---
label: e2e
---

## Expected

- Two observers both receive PTY output.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	p := resp.MultiAttachProbe
	if len(p.ObserverReceived) == 0 || len(p.Observer2Received) == 0 { t.Fatal("both observers should receive output") }
}
```
