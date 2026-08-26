## Expected

- No error.
- Only idA1 (title contains both MG_A and MG_B).
- idA2 / idB1 partial matches omitted.

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
