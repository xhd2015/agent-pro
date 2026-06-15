## Expected
- Single turn_start event.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"turn_start"`)
}
```
