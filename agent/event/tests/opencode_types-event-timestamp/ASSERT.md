## Expected
- Timestamp is parsed from JSON correctly.
- When re-marshaled, the timestamp is included in the output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "type=text")
	assertContains(t, resp.Output, "sessionID=sess_ts")
	assertContains(t, resp.Output, "timestamp=1718200000999")
	assertContains(t, resp.Output, "has_timestamp=true")
}
```
