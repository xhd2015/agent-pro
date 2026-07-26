## Expected
- Single tool_execution_update event.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_update"`)
	assertNotContains(t, resp.Output, `"type":"tool_execution_start"`)
	assertNotContains(t, resp.Output, `"type":"tool_execution_end"`)
}
```
