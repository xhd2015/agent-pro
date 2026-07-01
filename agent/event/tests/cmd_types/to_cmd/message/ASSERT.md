## Expected
- Output JSON array contains one event with `"role":"assistant"` and a `text` block.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"text":"Hello"`)
}
```
