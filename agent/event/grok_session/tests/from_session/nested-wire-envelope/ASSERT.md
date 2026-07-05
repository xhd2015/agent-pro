## Expected
- Same user ActionMessage as flat-user-chunk: role=user, text `run ls`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertEventCount(t, resp.Events, 1)
	ev := resp.Events[0]
	if ev.Type != types.ActionMessage || ev.Role != "user" || ev.Text != "run ls" {
		t.Fatalf("unexpected event: %#v", ev)
	}
}
```
