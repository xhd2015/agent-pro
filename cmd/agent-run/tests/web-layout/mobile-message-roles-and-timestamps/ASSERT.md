---
label: chromium
explanation: playwright; role bubble CSS distinction and message-timestamp visibility
---

## Expected

- `playwright-debug` exits 0.
- User and assistant bubbles use visually distinct alignment and/or background (checked in script).
- At least two `[data-testid="message-timestamp"]` elements with non-empty text.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright-debug exit %d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
	if req.Layout != "roles-timestamps" {
		t.Fatalf("expected layout roles-timestamps, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```