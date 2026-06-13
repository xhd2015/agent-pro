## Expected
- All five `ActionType` constants are printed with correct values.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "ActionThink=think")
	assertContains(t, resp.Stdout, "ActionToolCall=tool_call")
	assertContains(t, resp.Stdout, "ActionMessage=message")
	assertContains(t, resp.Stdout, "ActionError=error")
	assertContains(t, resp.Stdout, "ActionDone=done")
}
```
