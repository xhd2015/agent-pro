---
label: e2e
---

## Expected
- The command succeeds and prints deterministic model names.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    assertContains(t, resp.Stdout, "openai/gpt-5")
}
```

