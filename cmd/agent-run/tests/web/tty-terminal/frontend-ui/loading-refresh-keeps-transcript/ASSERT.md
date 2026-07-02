---
label: ui-automation
explanation: Uses browser automation to enforce the loading-with-existing-content rule.
---

## Expected

- Existing transcript remains visible during refresh.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
