## Expected
- Output JSON array contains one `result` event with `"subtype":"success"`, `"is_error":false`, and `"result":"ok"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"result"`)
	assertContains(t, resp.Output, `"subtype":"success"`)
	assertContains(t, resp.Output, `"is_error":false`)
	assertContains(t, resp.Output, `"result":"ok"`)
}
```
