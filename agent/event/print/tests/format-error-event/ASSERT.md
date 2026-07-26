## Expected
- Output is non-empty and contains FAILED (AgentEvent error primary path).
- Old opencode error events with nested error.data are now handled by AgentEvent
  primary path, which formats "type":"error" but does not extract nested error.message.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "FAILED")
}
```
