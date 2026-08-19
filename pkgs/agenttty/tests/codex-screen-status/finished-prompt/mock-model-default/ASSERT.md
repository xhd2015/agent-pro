## Expected

- `DetectScreenStatus` is `idle` (not `banner`).
- Fixture is live mock-model chrome (`default ·`, no stub banner).

## Exit Code

N/A (direct package call)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertScreenIdle(t, resp, err)
	if !strings.Contains(resp.Text, "default ·") {
		t.Fatalf("mock-model fixture must contain default · footer, got %q", resp.Text)
	}
}
```
