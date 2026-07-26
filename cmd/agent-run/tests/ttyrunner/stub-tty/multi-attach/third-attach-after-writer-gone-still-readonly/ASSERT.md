---
label: e2e
---

## Expected

- Third attach cannot write after writer detached.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ObserverCanWrite { t.Fatal("third attach must be read-only") }
}
```
