## Expected
- Output contains message_end type and the full assistant message with text and toolCall content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_end"`)
	assertContains(t, resp.Output, `"Hello world"`)
	assertContains(t, resp.Output, `"type":"toolCall"`)
	assertContains(t, resp.Output, `"id":"tc_1"`)
}
```
