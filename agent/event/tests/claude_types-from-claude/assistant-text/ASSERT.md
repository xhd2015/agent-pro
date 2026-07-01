## Expected
- Output JSON array contains one event with `"type":"message"` and text `pong`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"text":"pong"`)
}
```
