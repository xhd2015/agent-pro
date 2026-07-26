## Expected
- Output contains message_start type and user message text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_start"`)
	assertContains(t, resp.Output, `"role":"user"`)
	assertContains(t, resp.Output, `"Hello world"`)
}
```
