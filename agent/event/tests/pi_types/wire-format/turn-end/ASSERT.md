## Expected
- Output contains turn_end type, assistant message, and tool results.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"turn_end"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"role":"toolResult"`)
	assertContains(t, resp.Output, `"toolCallId":"call_1"`)
}
```
