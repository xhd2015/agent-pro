## Expected
- All JSON fields are present with correct values, including `tool_input`, `exit_code`, and `changes`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"id":"evt_001"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"text":"hello world"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"command":"echo hi"`)
	assertContains(t, resp.Output, `"output":"hi"`)
	assertContains(t, resp.Output, `"stderr":"err msg"`)
	assertContains(t, resp.Output, `"exit_code":42`)
	assertContains(t, resp.Output, `"path":"foo.txt"`)
	assertContains(t, resp.Output, `"kind":"add"`)
}
```
