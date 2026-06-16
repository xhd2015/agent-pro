## Expected
- PhaseStart (index 0): NOT skipped.
- PhaseUpdate (index 1): NOT skipped.
- PhaseEnd (index 2): SKIPPED.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 3 {
		t.Fatalf("expected 3 results, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("PhaseStart (empty) must not be skipped")
	}
	if resp.Skipped[1] {
		t.Fatalf("PhaseUpdate (empty) must not be skipped")
	}
	if !resp.Skipped[2] {
		t.Fatalf("PhaseEnd after start+update must be skipped (even if all empty)")
	}
}
```
