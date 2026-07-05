## Expected
- First turn events have `turn_index=0`; second turn events have `turn_index=1`.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	var turn0, turn1 []types.AgentEvent
	for _, ev := range resp.Events {
		switch grokTurnIndex(ev) {
		case 0:
			turn0 = append(turn0, ev)
		case 1:
			turn1 = append(turn1, ev)
		}
	}
	if len(turn0) == 0 || len(turn1) == 0 {
		t.Fatalf("expected events in both turns, got:\n%s", formatEvents(resp.Events))
	}
	assertHasTurnIndex(t, resp.Events, 0, 1)
}
```
