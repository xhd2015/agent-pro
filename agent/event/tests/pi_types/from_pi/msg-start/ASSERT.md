## Expected
- Single AgentEvent with ActionMessage and PhaseStart.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"phase":"start"`)
}
```
