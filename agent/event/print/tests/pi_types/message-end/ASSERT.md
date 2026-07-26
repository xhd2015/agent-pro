## Expected
- **After fix: message_end with no Delta produces empty output** (deltas already shown via message_update).
- No ASSISTANT marker and no "Bye" text should appear.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After fix: message_end without delta = empty output (no duplication)
	assertEmpty(t, resp.Output)
}
```
