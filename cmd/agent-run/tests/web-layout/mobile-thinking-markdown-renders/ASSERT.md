---
label: e2e, ui-automation
explanation: thinking progress card renders markdown structure
---

## Expected

- Playwright exit code **0**.
- Thinking card body has `strong` and/or `code` for seeded markdown markers.

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
	if req.Layout != "thinking-markdown-renders" {
		t.Fatalf("expected layout thinking-markdown-renders, got %q", req.Layout)
	}
}
```
