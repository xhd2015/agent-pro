## Expected
- PhaseStart must not be skipped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if len(resp.Skipped) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Skipped))
	}
	if resp.Skipped[0] {
		t.Fatalf("standalone PhaseStart must not be skipped")
	}
}
```
