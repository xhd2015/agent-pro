## Expected
- Output is empty array `[]` — non-assistant messages are skipped.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `[]`)
	assertNotContains(t, resp.Output, `"type"`)
}
```
