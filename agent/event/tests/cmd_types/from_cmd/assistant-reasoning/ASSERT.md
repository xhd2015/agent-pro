## Expected
- Output JSON array contains one event with `"type":"think"` and text `Let me think...`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"Let me think..."`)
}
```
