## Expected
- Output contains RUN label, the error message, and FAILED marker.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "not found")
	assertContains(t, resp.Output, "FAILED")
}
```
