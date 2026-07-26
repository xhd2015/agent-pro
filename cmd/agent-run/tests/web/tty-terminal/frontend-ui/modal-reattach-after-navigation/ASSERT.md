---
label: e2e, ui-automation
explanation: Uses browser automation to verify detach and reattach flow.
---

## Expected

- Second modal attach shows current terminal output for the same session.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
