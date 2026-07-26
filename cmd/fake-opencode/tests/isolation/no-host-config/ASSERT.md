---
label: e2e
---

## Expected
- No opencode config directory is created under temporary HOME.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.HostConfigExists {
        t.Fatal("fake-opencode wrote host config")
    }
}
```

