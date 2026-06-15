## Expected
- Single message_start event, no update or end.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertNotContains(t, resp.Output, `"type":"message_update"`)
	assertNotContains(t, resp.Output, `"type":"message_end"`)
}
```
