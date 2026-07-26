---
label: e2e, ui-automation
explanation: Uses browser automation to inspect the chat top bar.
---

## Expected

- A visible, enabled terminal button is present.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
