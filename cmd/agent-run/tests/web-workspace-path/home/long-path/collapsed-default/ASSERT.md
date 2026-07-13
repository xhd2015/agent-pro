---
label: ui-automation
explanation: playwright mobile home workspace collapsed default + runner viewport
---

## Expected

- `playwright-debug` exits 0.
- Viewport 390×844; no horizontal document scroll.
- `[data-testid="workspace"]` visible; label text is shortened (`…/last/two` via `shortWorkspaceLabel`).
- `title` attribute holds the full absolute server cwd.
- `[data-testid="runner-picker"]` / `runner-select` within viewport width.
- `[data-testid="empty-state"]` present (no seeded sessions).

## Side Effects

- Web process `cmd.Dir` = deep workspace for the leaf duration.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlaywrightOK(t, resp, err)
	if req.Scenario != "home-long-collapsed-default" {
		t.Fatalf("expected scenario home-long-collapsed-default, got %q", req.Scenario)
	}
	if req.WebWorkingDir == "" || req.WorkspacePath == "" {
		t.Fatal("WebWorkingDir/WorkspacePath must be set")
	}
}
```
