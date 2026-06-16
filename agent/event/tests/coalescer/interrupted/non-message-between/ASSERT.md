## Expected
- PhaseEnd (index 0): NOT skipped.
- ActionError (index 1): NOT skipped (non-message, always pass through).
- PhaseEnd (index 2): NOT skipped (state was reset by the error event).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 3 {
		t.Fatalf("expected 3 results, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("first PhaseEnd must not be skipped")
	}
	if resp.Skipped[1] {
		t.Fatalf("non-ActionMessage event must not be skipped")
	}
	if resp.Skipped[2] {
		t.Fatalf("PhaseEnd after state reset must not be skipped")
	}
}
```
