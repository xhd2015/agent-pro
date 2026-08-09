## Expected

- No error.
- Exactly one session: idFork (Kind `fork`).
- Whitespace forked_at session excluded.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertIDsAndKinds(t, resp.Sessions, idFork, "fork")
}
```
