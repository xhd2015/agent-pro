---
label: e2e, ui-automation
explanation: assistant message card renders markdown strong/pre/code
---

## Expected

- Playwright exit code **0**.
- Assistant message has `strong` and `pre`/`code`; no raw `**` / fence markers as sole text.

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
	if req.Layout != "assistant-markdown-renders" {
		t.Fatalf("expected layout assistant-markdown-renders, got %q", req.Layout)
	}
}
```
