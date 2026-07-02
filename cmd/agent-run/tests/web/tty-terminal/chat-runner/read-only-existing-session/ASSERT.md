---
label: ui-automation
explanation: Uses browser automation to inspect the existing chat page.
---

## Expected

- Runner text is visible.
- No enabled runner select exists on the chat page.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
