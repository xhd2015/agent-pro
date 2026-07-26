# Scenario

**Feature**: second tap collapses workspace path back to compact label on home

```
# expand then collapse
open / (deep) -> tap toggle (expand) -> tap toggle (collapse)
  -> label back to …/last/two; aria-expanded=false
  -> runner-picker still within viewport
```

## Preconditions

- Requires expand/collapse toggle (expect RED pre-impl).
- Deep workspace parent setup.

## Steps

1. Start web; open `/`.
2. Expand via `workspace-toggle`; verify full path.
3. Collapse via second tap; verify short label + runner viewport.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-long-tap-collapse"

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	full := jsString(req.WorkspacePath)
	expectedShort := jsString(shortWorkspaceLabel(req.WorkspacePath))

	body := clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const toggle = page.locator('[data-testid="workspace-toggle"]');
await toggle.waitFor({ state: 'visible', timeout: 15000 });
await toggle.click();
await page.waitForTimeout(100);
const expandedText = await workspaceLabelText();
if (expandedText !== '` + full + `') {
  throw new Error('expand step failed, got: ' + expandedText);
}
await toggle.click();
await page.waitForTimeout(100);
const collapsedText = await workspaceLabelText();
if (collapsedText !== '` + expectedShort + `') {
  throw new Error('expected collapsed short label ` + expectedShort + `, got: ' + collapsedText);
}
const expanded = await toggle.getAttribute('aria-expanded');
if (expanded !== 'false') {
  throw new Error('expected aria-expanded=false after collapse, got: ' + expanded);
}
` + assertRunnerPickerWithinViewport()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
