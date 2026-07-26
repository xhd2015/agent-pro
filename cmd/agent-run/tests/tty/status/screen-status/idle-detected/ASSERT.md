---
label: e2e
---

## Expected

- Exit code 0.
- Stdout contains "screen" with idle indicator (e.g., "idle" or "input prompt" visible).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "screen", "idle")
}
```
