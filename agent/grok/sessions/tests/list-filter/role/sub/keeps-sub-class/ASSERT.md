## Expected

- No error.
- Exactly four sessions newest-first:
  - idEmptyPar Kind `main` (sub class via parent; display default main)
  - idSubFork Kind `sub-f`
  - idSubRes Kind `sub+`
  - idSub Kind `sub`
- Plain main, fork, and empty-no-parent excluded.

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
		idEmptyPar, "main",
		idSubFork, "sub-f",
		idSubRes, "sub+",
		idSub, "sub",
	)
}
```
