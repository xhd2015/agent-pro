## Expected
- Three pi events: message_start, message_update (thinking_delta or thinking_start), message_end.
- All events reference the assistant role.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertContains(t, resp.Output, `"type":"message_end"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
}
```
