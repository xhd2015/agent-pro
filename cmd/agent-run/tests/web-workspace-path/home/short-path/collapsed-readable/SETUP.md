# Scenario

**Feature**: short workspace path remains readable on home without requiring expand

```
# short path regression
web cwd=short -> open / -> workspace label non-empty
  -> label text === shortWorkspaceLabel(full)
  -> runner-picker within viewport
```

## Preconditions

- Parent sets short `WebWorkingDir`.
- Expand is optional when full === short display; this leaf does not require toggle.

## Steps

1. Start web with short cwd.
2. Assert collapsed label matches `shortWorkspaceLabel` and runner stays visible.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-short-collapsed-readable"

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	expected := jsString(shortWorkspaceLabel(req.WorkspacePath))
	full := jsString(req.WorkspacePath)

	body := clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const text = await workspaceLabelText();
if (!text) throw new Error('expected readable workspace label, got empty');
if (text !== '` + expected + `') {
  throw new Error('expected shortWorkspaceLabel form ` + expected + `, got: ' + text);
}
const title = (await page.locator('[data-testid="workspace"]').getAttribute('title') || '').trim();
if (title && title !== '` + full + `') {
  throw new Error('title should be full path when set, got: ' + title);
}
` + assertRunnerPickerWithinViewport()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
