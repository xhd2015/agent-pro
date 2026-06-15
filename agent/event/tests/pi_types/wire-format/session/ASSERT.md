## Expected
- Marshaled output contains `"type":"session"` and `"id":"sess_abc123"`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"session"`)
	assertContains(t, resp.Output, `"id":"sess_abc123"`)
}
```
