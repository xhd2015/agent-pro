---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains JSON events from random generation (NOT the hardcoded "fake opencode answered" default).
- Events contain valid type and sessionID fields.

```go
import (
    "strings"
    "testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    if resp.Stdout == "" {
        t.Fatal("expected JSON events on stdout")
    }
    if strings.Contains(resp.Stdout, "fake opencode answered") {
        t.Fatal("output contains hardcoded default text, expected random generation")
    }
    assertContains(t, resp.Stdout, `"type":`)
    assertContains(t, resp.Stdout, `"sessionID":`)
}
```
