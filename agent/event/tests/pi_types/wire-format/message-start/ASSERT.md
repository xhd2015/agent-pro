## Expected
- Output contains message_start type and user message text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertContains(t, resp.Output, `"role":"user"`)
	assertContains(t, resp.Output, `"Hello world"`)
}
```
