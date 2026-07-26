---
label: e2e
---

## Expected

- Exit code 0 (status command succeeds even when screen can't be read).
- Stdout contains "screen" with "unknown" or error indicator.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "screen", "unknown")
}
```
