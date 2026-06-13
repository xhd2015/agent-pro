## Expected
- Two codex events: one `item.started` and one `item.completed` with item type `reasoning`.
- The completed event contains the think text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"item.started"`)
	assertContains(t, resp.Stdout, `"type":"item.completed"`)
	assertContains(t, resp.Stdout, `"type":"reasoning"`)
	assertContains(t, resp.Stdout, `"analyzing the request"`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
