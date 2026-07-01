## Expected
- Output JSON array contains one event with `"role":"assistant"` and a `tool-call` block.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"tool-call"`)
	assertContains(t, resp.Output, `"toolName":"bash"`)
	assertContains(t, resp.Output, `"command":"echo hi"`)
}
```
