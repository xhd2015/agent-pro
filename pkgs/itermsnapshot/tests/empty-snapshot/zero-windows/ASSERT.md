## Expected

- Capture succeeds with non-nil Result.Snapshot.
- Snapshot Windows empty; Agents nil or empty.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	res := mustResult(t, resp, err)
	if len(res.Snapshot.Windows) != 0 {
		t.Fatalf("Windows len=%d want 0", len(res.Snapshot.Windows))
	}
	if agentCount(res.Agents) != 0 {
		t.Fatalf("want 0 agents, got %d", agentCount(res.Agents))
	}
}
```
