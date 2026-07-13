---
label: ui-automation
explanation: mobile touch pan reveals older terminal LINE_* history
---

## Expected

- Playwright exit code **0**.
- After synthetic touch pan, min visible `LINE_*` index decreases (or equivalent scrollTop move into history).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "terminal-touch-scroll-history" {
		t.Fatalf("expected layout terminal-touch-scroll-history, got %q", req.Layout)
	}
}
```
