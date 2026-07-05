## Expected
- `SemanticEqual(events₁, events₂)`; user text and turn_index preserved.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSemanticEqualEvents(t, resp.Events1, resp.Events2)
	for _, ev := range resp.Events1 {
		if ev.Type == types.ActionMessage && ev.Role == "user" {
			if ev.Text != "run ls" {
				t.Fatalf("user text: got %q want run ls", ev.Text)
			}
			if grokTurnIndex(ev) != 0 {
				t.Fatalf("turn_index: got %d want 0", grokTurnIndex(ev))
			}
		}
	}
}
```
