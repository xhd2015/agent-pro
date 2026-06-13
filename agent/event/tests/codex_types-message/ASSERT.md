## Expected
- One codex event: `item.completed` with item type `message`.
- Contains the message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"type":"item.completed"`)
	assertContains(t, resp.Stdout, `"type":"message"`)
	assertContains(t, resp.Stdout, `"here is the result"`)
	assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
