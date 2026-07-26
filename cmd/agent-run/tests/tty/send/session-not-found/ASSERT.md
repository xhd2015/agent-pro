---
label: e2e
---

## Expected

- Exit code 1.
- Stderr mentions session not found or expired.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	assertOutput(t, resp, "stderr", "not found")
}
```
