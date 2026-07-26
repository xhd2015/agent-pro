## Expected
- Output contains isError:true and the error result.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_execution_end"`)
	assertContains(t, resp.Output, `"isError":true`)
	assertContains(t, resp.Output, `"file not found"`)
}
```
