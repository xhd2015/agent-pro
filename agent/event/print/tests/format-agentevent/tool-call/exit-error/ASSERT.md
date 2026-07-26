## Expected
- Output contains READ icon/label for read tool.
- Output contains the error message `file not found`.
- Output contains FAILED marker.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "READ")
	assertContains(t, resp.Output, "file not found")
	assertContains(t, resp.Output, "FAILED")
}
```
