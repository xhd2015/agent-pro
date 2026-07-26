## Expected
- JSON contains `"type":"error"`, `"sessionID":"sess_e1"`, and a nested `error` object with name and message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"error"`)
	assertContains(t, resp.Output, `"sessionID":"sess_e1"`)
	assertContains(t, resp.Output, `"name":"Error"`)
	assertContains(t, resp.Output, `"something went wrong"`)
}
```
