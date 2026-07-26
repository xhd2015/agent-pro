## Expected
- Second event has `status=failed`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertEventCount(t, resp.Events, 2)
	if grokStatus(resp.Events[1]) != "failed" {
		t.Fatalf("status: got %q want failed (%#v)", grokStatus(resp.Events[1]), resp.Events[1])
	}
}
```
