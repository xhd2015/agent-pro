## Expected
- Output is empty array `[]` — non-error finish produces no canonical action.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `[]`)
	assertNotContains(t, resp.Output, `"type"`)
}
```
