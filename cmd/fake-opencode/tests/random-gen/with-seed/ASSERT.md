---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains JSON events.
- Output is deterministic: contains known event types from seed 42.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.Stdout == "" {
        t.Fatal("expected JSON events on stdout")
    }
    assertContains(t, resp.Stdout, `"type":`)
    assertContains(t, resp.Stdout, `"sessionID":`)
}
```
