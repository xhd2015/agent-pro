## Expected
- Output is an empty JSON array (`[]`); no canonical action is emitted.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "[]")
	assertNotContains(t, resp.Output, `"type":"message"`)
	assertNotContains(t, resp.Output, `"type":"tool_call"`)
	assertNotContains(t, resp.Output, `"type":"done"`)
}
```
