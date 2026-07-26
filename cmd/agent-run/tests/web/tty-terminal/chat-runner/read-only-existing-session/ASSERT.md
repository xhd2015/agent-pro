---
label: e2e, ui-automation
explanation: Uses browser automation to inspect the existing chat page.
---

## Expected

- Runner text is visible.
- No enabled runner select exists on the chat page.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
