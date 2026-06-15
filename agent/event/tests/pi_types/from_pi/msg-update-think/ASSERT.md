## Expected
- Single AgentEvent with ActionThink and PhaseUpdate.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"phase":"update"`)
}
```
