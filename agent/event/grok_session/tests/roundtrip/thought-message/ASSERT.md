## Expected
- Think text preserved; semantic equality holds.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSemanticEqualEvents(t, resp.Events1, resp.Events2)
	found := false
	for _, ev := range resp.Events1 {
		if ev.Type == types.ActionThink && ev.Text == "planning ls output" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing think in events1:\n%s", formatEvents(resp.Events1))
	}
}
```
