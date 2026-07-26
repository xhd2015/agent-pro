---
label: e2e, ui-automation
explanation: Uses browser automation and websocket-backed terminal fixture.
---

## Expected

- Modal opens.
- Terminal output appears.
- Keyboard input including Enter is forwarded.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
