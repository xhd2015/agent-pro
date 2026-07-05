## Expected
- Two ActionDone events with turn_index 0 and 1; semantic equality.

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
	assertEventsOfType(t, resp.Events1, types.ActionDone, 2)
	assertHasTurnIndex(t, resp.Events1, 0, 1)
}
```
