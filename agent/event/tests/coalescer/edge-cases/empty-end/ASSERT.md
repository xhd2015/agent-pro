## Expected
- PhaseStart (index 0): NOT skipped.
- PhaseEnd (index 1): SKIPPED (even though end text is empty, start was shown).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("PhaseStart must not be skipped")
	}
	if !resp.Skipped[1] {
		t.Fatalf("PhaseEnd after PhaseStart must be skipped (even if end text is empty)")
	}
}
```
