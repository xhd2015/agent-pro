---
label: e2e, chromium, slow
explanation: Home poll + jump chip visibility and tap ~8s
---

## Expected

- Playwright exit code **0**.
- While detached with poll-added session (newest above), `[data-testid="jump-to-latest"]` becomes visible.
- After tap: `distanceFromTop <= 80`; chip no longer visible.

## Exit Code

- Playwright process exits 0.

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
	if req.Layout != "home-jump-to-latest" {
		t.Fatalf("expected layout home-jump-to-latest, got %q", req.Layout)
	}
}
```
