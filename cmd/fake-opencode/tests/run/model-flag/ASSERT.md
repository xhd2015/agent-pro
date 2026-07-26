---
label: e2e
---

## Expected
- The hook payload includes the model flag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.HookLog, `"model":"openai/gpt-5"`)
}
```

