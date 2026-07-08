---
label: chromium
explanation: playwright; grok-tty session message-card UX invariants
---

## Expected

- Playwright exits 0.
- At least 2 user and 2 assistant message cards with non-empty bodies.
- User vs assistant cards visually distinct; progress cards distinct from bubbles.
- No horizontal document overflow.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright exit=%d stderr=%s stdout=%s", resp.PlaywrightExit, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	if req.Layout != "grok-tty-message-cards" {
		t.Fatalf("expected layout grok-tty-message-cards, got %q", req.Layout)
	}
}
```