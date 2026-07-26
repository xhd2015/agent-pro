## Expected
- Output contains ASSISTANT marker and the reasoning text.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "ASSISTANT")
	assertContains(t, resp.Output, "Let me think about this carefully")
}
```
