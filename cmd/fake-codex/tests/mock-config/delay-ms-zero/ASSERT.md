---
label: e2e
---

## Expected
- The command succeeds quickly.

```go
import (
    "testing"
    "time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.Duration > 2*time.Second {
        t.Fatalf("duration = %s, want under 2s", resp.Duration)
    }
}
```

