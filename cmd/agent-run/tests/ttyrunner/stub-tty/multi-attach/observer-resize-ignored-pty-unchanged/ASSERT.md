---
label: e2e
---

## Expected

- Observer resize does not change PTY (resize dropped).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.MultiAttachProbe.ResizeAccepted { t.Fatal("observer resize must be ignored") }
}
```
