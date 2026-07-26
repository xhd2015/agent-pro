## Expected
- Output contains the 💬 icon.
- Empty text message still produces non-empty output.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output == "" {
		t.Fatalf("expected non-empty output for empty-text message")
	}
	assertContains(t, resp.Output, "💬")
}
```
