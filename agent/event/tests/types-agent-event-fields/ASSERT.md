## Expected
- All JSON fields are present with correct values, including `tool_input`, `exit_code`, and `changes`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, `"id":"evt_001"`)
	assertContains(t, resp.Stdout, `"type":"tool_call"`)
	assertContains(t, resp.Stdout, `"text":"hello world"`)
	assertContains(t, resp.Stdout, `"tool":"bash"`)
	assertContains(t, resp.Stdout, `"command":"echo hi"`)
	assertContains(t, resp.Stdout, `"output":"hi"`)
	assertContains(t, resp.Stdout, `"stderr":"err msg"`)
	assertContains(t, resp.Stdout, `"exit_code":42`)
	assertContains(t, resp.Stdout, `"path":"foo.txt"`)
	assertContains(t, resp.Stdout, `"kind":"add"`)
}
```
