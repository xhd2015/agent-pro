---
label: ui-automation
explanation: Uses browser automation and websocket-backed terminal fixture.
---

## Expected

- Modal opens.
- Terminal output appears.
- Keyboard input including Enter is forwarded.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
