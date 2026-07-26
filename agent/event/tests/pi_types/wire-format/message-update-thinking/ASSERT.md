## Expected
- Output contains thinking_delta event and the delta text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_update"`)
	assertContains(t, resp.Output, `"thinking_delta"`)
	assertContains(t, resp.Output, `"delta":" deeper"`)
}
```
