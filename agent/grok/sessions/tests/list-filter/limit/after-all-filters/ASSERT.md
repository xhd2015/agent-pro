## Expected

- No error.
- Exactly two sessions: idD1 then idA3 (newest place survivors).
- idB1 not included (failed place before limit).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertSessionIDs(t, resp.Sessions, idD1, idA3)
}
```
