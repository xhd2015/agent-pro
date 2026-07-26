---
label: e2e, ui-automation
explanation: playwright mobile home short workspace remains readable
---

## Expected

- Workspace label is non-empty and equals frontend `shortWorkspaceLabel(cwd)`.
- Runner-picker within viewport.
- No horizontal document scroll.
- Expand not required for this leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "home-short-collapsed-readable" {
		t.Fatalf("expected scenario home-short-collapsed-readable, got %q", req.Scenario)
	}
	if req.WorkspacePath == "" {
		t.Fatal("WorkspacePath must be set")
	}
}
```
