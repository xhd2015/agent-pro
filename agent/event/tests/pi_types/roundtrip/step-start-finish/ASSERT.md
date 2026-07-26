## Expected
- Roundtripped output preserves both step_start and step_finish actions.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"step_start"`)
	assertContains(t, resp.Output, `"type":"step_finish"`)
}
```
