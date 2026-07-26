## Expected
- Output JSON array contains three events in order: think, message, tool_call.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"thinking..."`)
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"text":"result"`)
	assertContains(t, resp.Output, `"type":"tool_call"`)
	assertContains(t, resp.Output, `"tool":"bash"`)
}
```
