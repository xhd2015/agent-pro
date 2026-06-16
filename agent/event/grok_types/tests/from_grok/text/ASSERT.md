## Expected
- One AgentEvent with type `message` and text matching the grok data.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"text":"Here is the response"`)
}
```
