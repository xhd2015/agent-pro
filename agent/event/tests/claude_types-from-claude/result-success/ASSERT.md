## Expected
- Output JSON array contains one event with `"type":"done"` and text `pong`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"done"`)
	assertContains(t, resp.Output, `"text":"pong"`)
	assertNotContains(t, resp.Output, `"type":"error"`)
}
```
