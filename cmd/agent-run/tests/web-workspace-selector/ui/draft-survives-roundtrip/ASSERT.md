---
label: ui-automation
explanation: Playwright composer draft survives /workspace Cancel round-trip
---

## Expected

- Playwright exit 0.
- After Cancel: composer input still contains typed draft.

## Errors

- Pre-impl: no /workspace route and/or local draft state wiped (RED).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
}
```
