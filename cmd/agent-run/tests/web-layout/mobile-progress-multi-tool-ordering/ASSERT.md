---
label: e2e, chromium
explanation: playwright; multi-tool progress compaction preserves start order
---

## Expected

- Playwright exits 0.
- Exactly 2 compacted `progress-card` elements.
- First card contains merged output from tool A; second card is tool B slot.

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
	if req.Layout != "progress-multi-tool-ordering" {
		t.Fatalf("expected layout progress-multi-tool-ordering, got %q", req.Layout)
	}
}
```