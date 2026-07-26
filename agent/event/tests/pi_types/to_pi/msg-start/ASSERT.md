## Expected
- Single message_start event, no update or end.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertNotContains(t, resp.Output, `"type":"message_update"`)
	assertNotContains(t, resp.Output, `"type":"message_end"`)
}
```
