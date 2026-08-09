## Expected

- No error.
- Exactly three sessions in newest-first order: idC1, idB1, idA1.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertSessionIDs(t, resp.Sessions, idC1, idB1, idA1)
}
```
