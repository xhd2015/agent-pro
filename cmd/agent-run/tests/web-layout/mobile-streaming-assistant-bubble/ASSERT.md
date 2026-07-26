---
label: e2e, chromium
explanation: Live web session + SSE/poll; assistant text growth ~20s
---

## Expected

- Playwright exit code **0** after detecting assistant bubble text growth during streaming.

## Exit Code

- Playwright process exits 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
}
```