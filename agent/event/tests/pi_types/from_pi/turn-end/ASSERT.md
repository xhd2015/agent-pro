## Expected
- Single AgentEvent with ActionStepFinish and PhaseEnd.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_finish"`)
	assertContains(t, resp.Output, `"phase":"end"`)
}
```
