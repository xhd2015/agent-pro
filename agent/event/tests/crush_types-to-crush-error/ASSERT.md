## Expected
- One crush event: type `agent_event` with nested type `error`.
- Contains the error text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"agent_event"`)
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"something went wrong"`)
}
```
