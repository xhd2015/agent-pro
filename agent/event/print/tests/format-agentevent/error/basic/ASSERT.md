## Expected
- Output contains the ❌ icon.
- Output contains the error text `boom`.
- Output contains FAILED marker.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "❌")
	assertContains(t, resp.Output, "boom")
	assertContains(t, resp.Output, "FAILED")
}
```
