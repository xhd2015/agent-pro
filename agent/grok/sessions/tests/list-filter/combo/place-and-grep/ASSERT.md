## Expected

- No error.
- Only idA1 (place + content). idB1 content match outside place excluded; idA2 no grep hit.

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
