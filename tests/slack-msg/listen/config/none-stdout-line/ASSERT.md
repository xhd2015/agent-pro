---
label: unit
explanation: daemon startup log probe
---

## Expected

- Startup output contains `Using config from: (none)`.
- Process stops cleanly on SIGTERM.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertOutputContains(t, resp, "Using config from: (none)")
}
```
