---
label: chromium, slow
explanation: Live stream + jump chip visibility and tap ~20s
---

## Expected

- Playwright exit code **0**.
- While detached with streaming growth, `[data-testid="jump-to-latest"]` becomes visible.
- After tap: `distanceFromBottom <= 80`; chip no longer visible.

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
	if req.Layout != "session-jump-to-latest" {
		t.Fatalf("expected layout session-jump-to-latest, got %q", req.Layout)
	}
}
```