## Expected

- No error.
- Zero sessions.
- Output exactly `No sessions found` (no table header).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertEmptyList(t, resp)
	assertNoSessionsFoundOutput(t, resp)
}
```
