## Expected

- Capture succeeds.
- Agents empty (idle sessions are never attached).

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
	if agentCount(res.Agents) != 0 {
		t.Fatalf("idle: want 0 agents, got %d", agentCount(res.Agents))
	}
}
```
