---
label: e2e, chromium
explanation: playwright; compacted progress cards distinct from chat bubbles
---

## Expected

- Playwright exits 0.
- Compacted progress card count is between 2 and 4 (duplicate tool_call rows collapsed).
- At least one `Thinking` and one `Tool` progress label visible.
- Progress card background differs from user message bubble.
- `.progress-card-body` has a max-height clamp.

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
	if req.Layout != "progress-card-compaction" {
		t.Fatalf("expected layout progress-card-compaction, got %q", req.Layout)
	}
}
```