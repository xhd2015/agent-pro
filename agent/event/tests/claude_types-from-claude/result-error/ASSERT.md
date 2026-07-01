## Expected
- Output JSON array contains one event with `"type":"error"` and text `boom`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"text":"boom"`)
	assertNotContains(t, resp.Output, `"type":"done"`)
}
```
