---
label: e2e
---

## Expected
- The `message.updated` hook fires with prompt text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"event":"message.updated"`)
    assertContains(t, resp.HookLog, `"text":"prompt text"`)
}
```

