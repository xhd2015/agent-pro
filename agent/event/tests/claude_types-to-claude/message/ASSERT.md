## Expected
- Output JSON array contains one `assistant` event with a `text` content block whose text is `pong`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"text":"pong"`)
}
```
