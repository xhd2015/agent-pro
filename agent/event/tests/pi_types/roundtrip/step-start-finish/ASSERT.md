## Expected
- Roundtripped output preserves both step_start and step_finish actions.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
	assertContains(t, resp.Output, `"type":"step_finish"`)
}
```
