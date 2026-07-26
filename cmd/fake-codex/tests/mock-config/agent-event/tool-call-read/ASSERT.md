---
label: e2e
---

## Expected
- The command succeeds.
- The completed codex event contains the file contents.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, `"type":"command_execution"`)
    assertContains(t, resp.Stdout, `"file contents here"`)
}
```
