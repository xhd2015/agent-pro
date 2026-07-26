## Expected
- Output is a JSON array with exactly one event of type `step_start`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
	assertNotContains(t, resp.Output, `"type":"message"`)
	assertNotContains(t, resp.Output, `"type":"done"`)
}
```
