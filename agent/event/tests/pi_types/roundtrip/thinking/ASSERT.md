## Expected
- Roundtripped output preserves ActionThink type and thinking text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"thinking deeply"`)
}
```
