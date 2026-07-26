---
label: e2e
---

## Expected

- Exit code 0.
- Output contains `from-config` from the sole config exchange.
- Output does not contain `from-events` (events file was not set).

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "from-config")
	assertNotContains(t, combined, "from-events")
}
```
