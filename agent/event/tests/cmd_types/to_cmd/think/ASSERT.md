## Expected
- Output JSON array contains one event with `"role":"assistant"` and a `reasoning` block.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"text":"thinking..."`)
}
```
