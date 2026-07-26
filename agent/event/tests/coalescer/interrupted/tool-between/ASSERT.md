## Expected
- PhaseEnd (index 0): NOT skipped.
- ActionToolCall (index 1): NOT skipped.
- PhaseEnd (index 2): NOT skipped (state was reset by tool call).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 3 {
		t.Fatalf("expected 3 results, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("first PhaseEnd must not be skipped")
	}
	if resp.Skipped[1] {
		t.Fatalf("ActionToolCall must not be skipped")
	}
	if resp.Skipped[2] {
		t.Fatalf("PhaseEnd after tool call (state reset) must not be skipped")
	}
}
```
