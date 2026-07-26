---
label: e2e
---

## Expected
- The command succeeds.
- stdout contains no events (empty or only newlines).

```go
import (
	"testing"
	"strings"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    trimmed := strings.TrimSpace(resp.Stdout)
    if trimmed != "" {
        t.Fatalf("expected empty stdout, got:\n%s", resp.Stdout)
    }
}
```
