## Expected
- Roundtripped output preserves `type:think` and thinking text content.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"Analyzing the request..."`)
}
```
