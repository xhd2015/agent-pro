---
label: e2e, ui-automation
explanation: playwright session page workspace collapsed + tap expand (expect RED pre-impl)
---

## Expected

- Session header `[data-testid="workspace"]` uses compact short label by default.
- Tap `[data-testid="workspace-toggle"]` shows full `meta.workspace` path in label.
- `aria-expanded="true"` after expand.
- No horizontal document scroll.
- `[data-testid="chat-active"]` visible.

## Errors

- Pre-impl: session shows raw full path (no short / no toggle) → RED.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "session-long-tap-expand" {
		t.Fatalf("expected scenario session-long-tap-expand, got %q", req.Scenario)
	}
	if req.WorkspacePath == "" || req.SessionID == "" {
		t.Fatal("WorkspacePath and SessionID must be set")
	}
}
```
