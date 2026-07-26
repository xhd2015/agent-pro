## Expected
- One canonical AgentEvent with type `tool_call`.
- Tool name is `bash`, tool_input contains the command.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"command"`)
	assertContains(t, resp.Output, `"echo hi"`)
}
```
