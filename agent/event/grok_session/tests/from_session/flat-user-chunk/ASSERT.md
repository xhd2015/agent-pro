## Expected
- One `ActionMessage` with `role=user`, text `run ls`, `turn_index=0`.

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
	if got := grokTurnIndex(ev); got != 0 {
		t.Fatalf("turn_index: got %d want 0", got)
	}
}
```
