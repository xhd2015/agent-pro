## Expected

- No error.
- Zero sessions; `No sessions found`.

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
