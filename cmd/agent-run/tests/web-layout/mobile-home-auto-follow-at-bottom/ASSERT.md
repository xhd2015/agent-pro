---
label: chromium, slow
explanation: Home poll refresh ~4s; follow-at-bottom distance checks
---

## Expected

- Playwright exit code **0**.
- Initial `distanceFromBottom <= 80` on `session-list`.
- After poll refresh adds a 21st session, `distanceFromBottom` stays `<= 80`.

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
	if req.Layout != "home-auto-follow-at-bottom" {
		t.Fatalf("expected layout home-auto-follow-at-bottom, got %q", req.Layout)
	}
}
```