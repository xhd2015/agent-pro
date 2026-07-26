## Expected
- One canonical AgentEvent with type `done`.
- Text contains the run output text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"success output"`)
}
```
