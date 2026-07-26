## Expected
- One canonical AgentEvent with type `done`.
- Text contains the error message from run_complete.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"agent run failed"`)
}
```
