## Expected
- One codex event: `item.completed` with item type `message`.
- Contains the message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"item.completed"`)
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"here is the result"`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
