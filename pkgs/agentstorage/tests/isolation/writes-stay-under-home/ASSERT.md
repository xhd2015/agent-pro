## Expected

- At least one file was written under home.
- Every path in `FilesWritten` is prefixed by `ResolvedHome`.
- No files appear outside the temp home directory.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if len(resp.FilesWritten) == 0 {
		t.Fatal("expected files written during isolation test")
	}
	AssertHomeOnly(t, resp.ResolvedHome, resp.FilesWritten)
}
```