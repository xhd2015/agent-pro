---
label: e2e
---

## Expected
- The tool event has the correct shape: type `tool_use`, tool `bash`, status `completed`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"bash"`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
