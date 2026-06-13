## Expected
- All five `ActionType` constants are printed with correct values.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "ActionThink=think")
	assertContains(t, resp.Output, "ActionToolCall=tool_call")
	assertContains(t, resp.Output, "ActionMessage=message")
	assertContains(t, resp.Output, "ActionError=error")
	assertContains(t, resp.Output, "ActionDone=done")
}
```
