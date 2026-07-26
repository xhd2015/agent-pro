## Expected
- JSON contains `"type":"tool_use"`, `"sessionID":"sess_t1"`, and a `part` with tool name, state, output, exit code.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_use"`)
	assertContains(t, resp.Output, `"sessionID":"sess_t1"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"output":"hello"`)
	assertContains(t, resp.Output, `"exit_code":0`)
	assertContains(t, resp.Output, `"status":"completed"`)
}
```
