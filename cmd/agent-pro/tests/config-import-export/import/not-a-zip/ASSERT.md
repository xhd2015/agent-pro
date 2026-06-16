## Expected
- An error is returned because the file is not a valid zip.
- No files are written to disk.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertError(t, resp)
}
```
