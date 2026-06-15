## Expected
- Two pi events: message_start and message_end.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertContains(t, resp.Output, `"type":"message_end"`)
}
```
