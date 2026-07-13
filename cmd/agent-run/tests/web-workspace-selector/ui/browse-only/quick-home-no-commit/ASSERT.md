---
label: ui-automation
explanation: Playwright Quick Home chip browse-only — no status commit
---

## Expected

- Playwright exit 0.
- After Quick Home: browse path updated; `status.workspace` still selected; still on `/workspace`.

## Errors

- Pre-impl: selector/chip missing or chip auto-commits (RED).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
