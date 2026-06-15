## Expected
- Single AgentEvent with ActionDone and PhaseEnd.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"phase":"end"`)
}
```
