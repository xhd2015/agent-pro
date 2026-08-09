## Expected

- No error.
- Five sessions newest-first with Kind: fork, sub-f, sub+, sub, main.

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
		idFork, "fork",
		idSubFork, "sub-f",
		idSubRes, "sub+",
		idSub, "sub",
		idMain, "main",
	)
}
```
