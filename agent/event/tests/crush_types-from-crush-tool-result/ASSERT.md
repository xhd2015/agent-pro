## Expected
- Output is empty array `[]` — tool_result parts produce no canonical actions.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `[]`)
	assertNotContains(t, resp.Output, `"type"`)
}
```
