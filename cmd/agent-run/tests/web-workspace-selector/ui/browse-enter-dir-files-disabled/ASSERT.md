---
label: ui-automation
explanation: Playwright expands files then asserts file disabled; enter dir updates path
---

## Expected

- Playwright exit 0.
- After expand: file entry non-navigable; entering `subdir` updates browse path.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
