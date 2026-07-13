---
label: ui-automation
explanation: Playwright lists and enters dot directories (e.g. .hidden-dir)
---

## Expected

- Playwright exit 0.
- `.hidden-dir` visible as dir; enter updates browse path.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
