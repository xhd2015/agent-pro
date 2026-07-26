---
label: e2e, ui-automation
explanation: Uses browser automation to enforce the loading-with-existing-content rule.
---

## Expected

- Existing transcript remains visible during refresh.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
