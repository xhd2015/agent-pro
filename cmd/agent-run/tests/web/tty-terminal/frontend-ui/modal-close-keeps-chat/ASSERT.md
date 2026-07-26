---
label: e2e, ui-automation
explanation: Uses browser automation to verify modal close behavior.
---

## Expected

- Closing modal hides only the modal.
- Chat transcript remains visible.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
