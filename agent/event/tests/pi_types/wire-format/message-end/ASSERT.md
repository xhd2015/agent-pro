## Expected
- Output contains message_end type and the full assistant message with text and toolCall content.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_end"`)
	assertContains(t, resp.Output, `"Hello world"`)
	assertContains(t, resp.Output, `"type":"toolCall"`)
	assertContains(t, resp.Output, `"id":"tc_1"`)
}
```
