## Expected
- The event must not be skipped (index 0 = false).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("standalone PhaseEnd must not be skipped")
	}
}
```
