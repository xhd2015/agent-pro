## Expected
- Output JSON array contains one `result` event with `"subtype":"error"`, `"is_error":true`, and `"result":"boom"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"result"`)
	assertContains(t, resp.Output, `"subtype":"error"`)
	assertContains(t, resp.Output, `"is_error":true`)
	assertContains(t, resp.Output, `"result":"boom"`)
}
```
