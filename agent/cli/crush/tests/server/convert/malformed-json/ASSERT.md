## Expected
- `UnwrapEvent` returns `nil, nil` (malformed JSON dropped gracefully, no error).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("convert failed: %v", err)
	}
	if resp.Output != "null" {
		t.Fatalf("expected Output %q for malformed input, got %q", "null", resp.Output)
	}
}
```
