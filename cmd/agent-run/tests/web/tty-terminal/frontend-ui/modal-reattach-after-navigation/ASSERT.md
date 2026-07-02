---
label: ui-automation
explanation: Uses browser automation to verify detach and reattach flow.
---

## Expected

- Second modal attach shows current terminal output for the same session.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
