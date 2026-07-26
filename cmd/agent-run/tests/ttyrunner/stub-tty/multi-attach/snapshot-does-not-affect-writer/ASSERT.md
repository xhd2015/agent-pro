---
label: e2e
---

## Expected

- Writer can still write after snapshot attach.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.WriterCanWrite { t.Fatal("writer should retain write after snapshot probe") }
}
```
