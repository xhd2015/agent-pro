## Expected
- One AgentEvent with type `think` and text matching the grok thought data.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"I need to consider the user's request"`)
}
```
