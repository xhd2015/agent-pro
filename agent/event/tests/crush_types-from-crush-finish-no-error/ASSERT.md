## Expected
- Output is empty array `[]` — non-error finish produces no canonical action.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `[]`)
	assertNotContains(t, resp.Output, `"type"`)
}
```
