---
label: e2e
---

## Expected

- After writer detach, earlier second attach cannot write.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ObserverCanWrite { t.Fatal("second attach must remain read-only after writer detach") }
}
```
