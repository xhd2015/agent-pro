## Expected
- Output contains message_update type, text_delta event, and delta content.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message_update"`)
	assertContains(t, resp.Output, `"text_delta"`)
	assertContains(t, resp.Output, `"delta":"lo"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
}
```
