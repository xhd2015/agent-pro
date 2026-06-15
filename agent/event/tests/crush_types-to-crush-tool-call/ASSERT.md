## Expected
- One crush event: type `message` with role `assistant`.
- Message parts contain a `tool_call` part with name `bash` and input JSON.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"name":"bash"`)
	assertContains(t, resp.Output, `"echo hello"`)
}
```
