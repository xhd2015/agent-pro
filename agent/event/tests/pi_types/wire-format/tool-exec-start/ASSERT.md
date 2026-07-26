## Expected
- Output contains tool_execution_start type, toolCallId, toolName, and args.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_start"`)
	assertContains(t, resp.Output, `"toolCallId":"call_1"`)
	assertContains(t, resp.Output, `"toolName":"bash"`)
	assertContains(t, resp.Output, `"ls -la"`)
}
```
