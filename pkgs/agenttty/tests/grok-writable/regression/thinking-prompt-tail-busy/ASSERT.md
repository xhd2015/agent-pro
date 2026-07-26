## Expected

- `ready=false`, `state=busy`.
- `reason` contains agent-still-responding semantics (non-empty).

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertWritable(t, "thinking busy", resp.Status, false, "busy", "agent still responding")
}
```