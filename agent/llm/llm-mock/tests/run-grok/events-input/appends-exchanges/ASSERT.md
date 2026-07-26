---
label: e2e
---

## Expected

- Exit code 0.
- Combined output contains `from-config` (first curl) and `from-events` (second curl from events file).

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
	assertContains(t, combined, "from-events")
}
```
