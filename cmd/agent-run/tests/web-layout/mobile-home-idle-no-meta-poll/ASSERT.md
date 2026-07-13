---
label: ui-automation, slow
explanation: ~7s idle home poll hygiene (no runners/status delta)
---

## Expected

- Playwright exit code **0**.
- After bootstrap, idle 7s: runners GET Δ=0, status GET Δ=0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "home-idle-no-meta-poll" {
		t.Fatalf("expected layout home-idle-no-meta-poll, got %q", req.Layout)
	}
}
```
