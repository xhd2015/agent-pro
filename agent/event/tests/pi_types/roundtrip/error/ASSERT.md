## Expected
- Roundtripped output preserves ActionError type and error message.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"something broke"`)
}
```
