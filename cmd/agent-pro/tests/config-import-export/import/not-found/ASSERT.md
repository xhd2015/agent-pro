## Expected
- An error is returned because the zip file does not exist.
- No files are written to disk.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertError(t, resp)
}
```
