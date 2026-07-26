## Expected
- Emits user message then `ActionDone` with `turn_index=0`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertEventsOfType(t, resp.Events, types.ActionDone, 1)
	done := resp.Events[len(resp.Events)-1]
	if done.Type != types.ActionDone {
		t.Fatalf("last event not done: %#v", done)
	}
	if got := grokTurnIndex(done); got != 0 {
		t.Fatalf("done turn_index: got %d want 0", got)
	}
}
```
