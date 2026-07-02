---
label: ui-automation
explanation: Uses browser automation to verify modal close behavior.
---

## Expected

- Closing modal hides only the modal.
- Chat transcript remains visible.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
