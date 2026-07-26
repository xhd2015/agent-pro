---
label: e2e
---

## Expected

- `tty send` succeeds via server write while client holds unified write.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.SendInjected { t.Fatal("expected server-side send to inject while writer attached") }
}
```
