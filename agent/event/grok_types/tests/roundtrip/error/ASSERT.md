## Expected
- Roundtripped output preserves `type:error` and the error message text.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"text":"connection refused"`)
}
```
