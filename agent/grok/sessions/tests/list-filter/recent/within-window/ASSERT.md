## Expected

- No error.
- Only idA1; idB1 outside window excluded.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertSessionIDs(t, resp.Sessions, idA1)
}
```
