---
label: e2e, ui-automation
explanation: playwright mobile home workspace tap-to-expand full path (expect RED pre-impl)
---

## Expected

- `playwright-debug` exits 0 (after feature lands).
- Before tap: shortened `…/` label.
- After tap `[data-testid="workspace-toggle"]`:
  - `[data-testid="workspace-label"]` text equals full absolute path
  - `aria-expanded="true"` on toggle
- No horizontal document scroll when expanded.

## Errors

- Today (pre-impl): missing `workspace-toggle` / expand behavior → playwright non-zero (RED).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "home-long-tap-expand" {
		t.Fatalf("expected scenario home-long-tap-expand, got %q", req.Scenario)
	}
	if req.WorkspacePath == "" {
		t.Fatal("WorkspacePath must be set")
	}
}
```
