## Expected
- Output contains the ❌ icon.
- Output contains FAILED marker.
- Empty-text error still produces non-empty output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output == "" {
		t.Fatalf("expected non-empty output for empty-text error")
	}
	assertContains(t, resp.Output, "❌")
	assertContains(t, resp.Output, "FAILED")
}
```
