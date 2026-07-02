---
label: ui-automation
explanation: Uses browser automation to inspect the chat top bar.
---

## Expected

- A visible, enabled terminal button is present.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
