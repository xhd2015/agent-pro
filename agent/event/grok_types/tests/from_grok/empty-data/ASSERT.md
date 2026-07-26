## Expected
- One AgentEvent with type `message` and empty text (or no text field).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertNotContains(t, resp.Output, `"text":"`)
}
```
