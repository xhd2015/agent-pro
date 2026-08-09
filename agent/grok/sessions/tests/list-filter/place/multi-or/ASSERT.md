## Expected

- No error.
- Exactly two sessions newest-first among survivors: idB1 then idA1.
- idC1 excluded.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertSessionIDs(t, resp.Sessions, idB1, idA1)
}
```
