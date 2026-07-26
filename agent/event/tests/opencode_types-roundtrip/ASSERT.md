## Expected
- Every opencode event type round-trips unchanged: the output must contain `ALL MATCH`.
- If any type mismatches, the output shows `[FAIL] <type>` with orig and rt JSON for diagnosis.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "ALL MATCH")
}
```
