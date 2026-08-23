## Expected

- Capture succeeds (resolve error is soft).
- Agents empty.

## Errors

- `err` is nil (hard error only from base capture, not resolve).

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
	if agentCount(res.Agents) != 0 {
		t.Fatalf("resolve error: want 0 agents, got %d", agentCount(res.Agents))
	}
}
```
