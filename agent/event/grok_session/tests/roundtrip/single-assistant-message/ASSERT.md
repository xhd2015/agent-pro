## Expected
- Assistant text preserved across roundtrip.

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
	found := false
	for _, ev := range resp.Events1 {
		if ev.Type == types.ActionMessage && ev.Role == "assistant" && ev.Text == "Here is the answer" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing assistant message in events1:\n%s", formatEvents(resp.Events1))
	}
}
```
