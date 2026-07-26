## Expected
- Roundtripped output preserves ActionThink type and thinking text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"thinking deeply"`)
}
```
