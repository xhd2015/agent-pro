## Expected

- No error.
- Exactly three sessions, newest-first: idEmptyNo (Kind `main`), idFork (Kind `fork`), idMain (Kind `main`).
- All sub-agent class fixtures excluded.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertIDsAndKinds(t, resp.Sessions,
		idEmptyNo, "main",
		idFork, "fork",
		idMain, "main",
	)
}
```
