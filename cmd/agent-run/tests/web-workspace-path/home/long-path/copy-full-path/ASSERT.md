---
label: e2e, ui-automation
explanation: playwright clipboard copy of full workspace path (expect RED pre-impl)
---

## Expected

- After expand, `[data-testid="workspace-copy"]` is visible and clickable.
- `navigator.clipboard.readText()` equals the full absolute workspace path.
- No horizontal document scroll.

## Errors

- Pre-impl: missing copy control / toggle → RED.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "home-long-copy-full-path" {
		t.Fatalf("expected scenario home-long-copy-full-path, got %q", req.Scenario)
	}
	if req.WorkspacePath == "" {
		t.Fatal("WorkspacePath must be set")
	}
}
```
