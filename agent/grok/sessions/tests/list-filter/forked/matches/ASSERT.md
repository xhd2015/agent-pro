## Expected

- No error.
- Three sessions newest-first: idForkedAt (Kind `main`), idSubFork (Kind `sub-f`), idFork (Kind `fork`).
- Plain main excluded.

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
		idForkedAt, "main",
		idSubFork, "sub-f",
		idFork, "fork",
	)
}
```
