## Expected
- Single AgentEvent with ActionMessage and PhaseEnd.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"phase":"end"`)
}
```
