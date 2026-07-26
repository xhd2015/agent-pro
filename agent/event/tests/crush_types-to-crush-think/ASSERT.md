## Expected
- One crush event: payload type `message` with role `assistant`.
- Message parts contain a `reasoning` part with the think text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"type":"reasoning"`)
	assertContains(t, resp.Output, `"thinking about the problem"`)
}
```
