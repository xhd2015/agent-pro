---
label: e2e
---

## Expected

- Registry file still exists after observer disconnect.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if !resp.MultiAttachProbe.SessionStillAlive { t.Fatal("session should remain after observer disconnect") }
}
```
