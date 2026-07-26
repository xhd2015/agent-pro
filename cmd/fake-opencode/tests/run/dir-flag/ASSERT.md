---
label: e2e
---

## Expected
- The command accepts `--dir`.
- Host opencode config is not created.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.HostConfigExists {
        t.Fatal("fake-opencode wrote host opencode config")
    }
}
```

