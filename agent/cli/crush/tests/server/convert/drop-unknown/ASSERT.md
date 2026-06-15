## Expected
- `UnwrapEvent` returns `nil, nil` (event dropped, no error).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if resp.Output != "null" {
		t.Fatalf("expected Output %q for dropped event, got %q", "null", resp.Output)
	}
}
```
