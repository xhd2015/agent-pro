---
label: ui-automation
explanation: Uses browser automation to inspect runner picker options.
---

## Expected

- Home runner select contains both tty runner options.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
