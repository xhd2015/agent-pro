---
label: e2e
---

## Expected

- Writer accepts input; second attach input not echoed (read-only).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	p := resp.MultiAttachProbe
	if p == nil { t.Fatal("nil probe") }
	if !p.WriterCanWrite { t.Fatal("writer should accept input") }
	if p.ObserverCanWrite { t.Fatal("second attach must not echo observer input") }
}
```
