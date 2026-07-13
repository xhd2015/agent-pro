---
label: ui-automation, slow
explanation: multi-step home scroll settle (~3.5s) must not snap back
---

## Expected

- Playwright exit code **0**.
- After A→B→C settle, `session-list.scrollTop` stays near C (±80px), not snapped to B.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "home-scroll-no-snap-back" {
		t.Fatalf("expected layout home-scroll-no-snap-back, got %q", req.Layout)
	}
}
```
