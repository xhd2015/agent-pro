## Expected
- Two codex events: one `item.started` and one `item.completed` with item type `reasoning`.
- The completed event contains the think text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"item.started"`)
	assertContains(t, resp.Output, `"type":"item.completed"`)
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"analyzing the request"`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
