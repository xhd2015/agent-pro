## Expected
- Output JSON array contains one event with `"type":"think"` and text `Let me think...`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"text":"Let me think..."`)
}
```
