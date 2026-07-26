## Expected
- Emits user, think, tool (pending+completed), assistant, and ActionDone.

```go
import (
	"testing"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, ev := range resp.Events {
		switch ev.Type {
		case types.ActionMessage:
			if ev.Role == "user" {
				found["user"] = true
			}
			if ev.Role == "assistant" {
				found["assistant"] = true
			}
		case types.ActionThink:
			found["think"] = true
		case types.ActionToolCall:
			found["tool"] = true
		case types.ActionDone:
			found["done"] = true
		}
	}
	for _, key := range []string{"user", "think", "tool", "assistant", "done"} {
		if !found[key] {
			t.Fatalf("missing %s in sequence:\n%s", key, formatEvents(resp.Events))
		}
	}
}
```
