## Expected
- Output contains message_update type, text_delta event, and delta content.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_update"`)
	assertContains(t, resp.Output, `"text_delta"`)
	assertContains(t, resp.Output, `"delta":"lo"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
}
```
