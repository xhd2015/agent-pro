---
label: ui-automation
explanation: Playwright chat DOM poll for assistant bubble count
---

## Expected

- Playwright exits 0 with exactly one assistant bubble and no phased DOM rows.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	requirePlaywrightOK(t, resp, err)
}
```
