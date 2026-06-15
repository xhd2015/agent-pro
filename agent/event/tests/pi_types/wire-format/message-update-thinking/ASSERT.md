## Expected
- Output contains thinking_delta event and the delta text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_update"`)
	assertContains(t, resp.Output, `"thinking_delta"`)
	assertContains(t, resp.Output, `"delta":" deeper"`)
}
```
