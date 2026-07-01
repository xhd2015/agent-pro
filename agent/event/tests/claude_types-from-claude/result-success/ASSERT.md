## Expected
- Output JSON array contains one event with `"type":"done"` and text `pong`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"text":"pong"`)
	assertNotContains(t, resp.Output, `"type":"error"`)
}
```
