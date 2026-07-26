---
label: e2e
---

## Expected
- The command succeeds.
- The tool_use event contains the mocked output, stderr, and exit code.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"tool_use"`)
    assertContains(t, resp.Stdout, `"tool":"bash"`)
    assertContains(t, resp.Stdout, `"mocked-stdout"`)
    assertContains(t, resp.Stdout, `"mocked-stderr"`)
    assertContains(t, resp.Stdout, `"exit_code":1`)
    assertContains(t, resp.Stdout, `"status":"completed"`)
}
```
