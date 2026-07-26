## Expected
- One crush event: type `message` with role `assistant`.
- Message parts contain a `text` part with the message text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"here is the result"`)
}
```
