## Expected
- `step_start` AgentEvent serializes with `"type":"step_start"` and timestamp.
- `step_finish` AgentEvent serializes with `"type":"step_finish"` and timestamp.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
	assertContains(t, resp.Output, `"type":"step_finish"`)
	assertContains(t, resp.Output, `"timestamp":1718200000123`)
	assertContains(t, resp.Output, `"timestamp":1718200000456`)
}
```
