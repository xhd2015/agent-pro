## Expected
- Output JSON array contains one ActionToolCall with `"output":"hi"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
	assertContains(t, resp.Output, `"output":"hi"`)
}
```
