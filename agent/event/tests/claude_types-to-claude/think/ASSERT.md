## Expected
- Output JSON array contains one `assistant` event with a `thinking` content block whose `thinking` text is `reasoning`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"type":"thinking"`)
	assertContains(t, resp.Output, `"thinking":"reasoning"`)
}
```
