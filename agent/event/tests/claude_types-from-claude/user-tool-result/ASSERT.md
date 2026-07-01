## Expected
- Output is an empty JSON array (`[]`); no canonical action is emitted.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "[]")
	assertNotContains(t, resp.Output, `"type":"message"`)
	assertNotContains(t, resp.Output, `"type":"tool_call"`)
	assertNotContains(t, resp.Output, `"type":"done"`)
}
```
