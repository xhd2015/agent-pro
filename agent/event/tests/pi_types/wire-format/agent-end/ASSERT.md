## Expected
- Output contains `"type":"agent_end"` and the messages array.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"agent_end"`)
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, `"role":"user"`)
	assertContains(t, resp.Output, `"Hello"`)
}
```
