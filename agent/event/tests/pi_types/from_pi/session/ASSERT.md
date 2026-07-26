## Expected
- Empty result: no AgentEvent produced from session event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.Output != "null" && resp.Output != "[]" && resp.Output != "" {
		t.Fatalf("expected empty output, got: %s", resp.Output)
	}
}
```
