---
label: chromium
explanation: playwright detach + composer send scroll position
---

## Expected

- Playwright exit code **0**.
- After scroll-up detach, `distanceFromBottom > 80`.
- After composer send, `message-list.scrollTop` unchanged (±2px).

## Exit Code

- Playwright process exits 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "session-send-no-auto-scroll" {
		t.Fatalf("expected layout session-send-no-auto-scroll, got %q", req.Layout)
	}
}
```