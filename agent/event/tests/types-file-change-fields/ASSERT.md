## Expected
- JSON contains `path` and `kind` fields.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"path":"bar.go"`)
	assertContains(t, resp.Output, `"kind":"modify"`)
}
```
