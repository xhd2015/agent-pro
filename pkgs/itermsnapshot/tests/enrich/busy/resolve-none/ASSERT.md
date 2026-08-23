## Expected

- Capture succeeds (soft miss).
- Agents empty — Kind none does not produce SessionAgent.

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
		t.Fatalf("Kind=none: want 0 agents, got %d", agentCount(res.Agents))
	}
}
```
