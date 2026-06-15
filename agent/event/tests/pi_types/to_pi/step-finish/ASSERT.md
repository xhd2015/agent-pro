## Expected
- Single turn_end event.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"turn_end"`)
}
```
