## Expected
- Output JSON array contains one `assistant` event with a `tool_use` block: `name="Bash"`, `input={"command":"ls"}`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"type":"tool_use"`)
	assertContains(t, resp.Output, `"name":"Bash"`)
	assertContains(t, resp.Output, `"input":{"command":"ls"}`)
}
```
