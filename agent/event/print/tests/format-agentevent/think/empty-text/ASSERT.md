## Expected
- Output contains the 💭 icon.
- Empty think still produces non-empty output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Output == "" {
		t.Fatalf("expected non-empty output for empty-text think")
	}
	assertContains(t, resp.Output, "💭")
}
```
