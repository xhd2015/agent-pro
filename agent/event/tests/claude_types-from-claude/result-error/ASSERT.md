## Expected
- Output JSON array contains one event with `"type":"error"` and text `boom`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"text":"boom"`)
	assertNotContains(t, resp.Output, `"type":"done"`)
}
```
