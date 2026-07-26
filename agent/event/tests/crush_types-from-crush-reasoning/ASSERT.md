## Expected
- One canonical AgentEvent with type `think`.
- Text contains the reasoning content.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"think"`)
	assertContains(t, resp.Output, `"step by step reasoning"`)
}
```
