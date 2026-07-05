## Expected
- Failed status preserved across roundtrip.

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
		if ev.Type == types.ActionToolCall && grokStatus(ev) == "failed" {
			return
		}
	}
	t.Fatal("missing failed tool event")
}
```
