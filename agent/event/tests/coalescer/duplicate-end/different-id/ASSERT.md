## Expected
- Both PhaseEnd events (index 0, 1): NOT skipped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("first PhaseEnd (ID=a) must not be skipped")
	}
	if resp.Skipped[1] {
		t.Fatalf("second PhaseEnd (ID=b, different ID) must not be skipped")
	}
}
```
