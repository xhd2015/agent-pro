## Expected

- Every emitted event has `extensions.grok_session.turn_index=0`.
- Exactly one `ActionDone` event is present (from `turn_completed`).

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Events) < 5 {
		t.Fatalf("expected at least 5 events, got %d:\n%s", len(resp.Events), formatEvents(resp.Events))
	}
	assertAllTurnIndex(t, resp.Events, 0)
	assertEventsOfType(t, resp.Events, types.ActionDone, 1)
	done := resp.Events[len(resp.Events)-1]
	if done.Type != types.ActionDone {
		t.Fatalf("last event not ActionDone: %#v", done)
	}
	if got := grokTurnIndex(done); got != 0 {
		t.Fatalf("ActionDone turn_index: got %d want 0", got)
	}
}
```