# Scenario

**Feature**: long workspace on home defaults to compact label; runner stays visible

```
# default collapsed state (no tap)
web cwd=deep -> open / -> workspace shows …/last/two
  -> runner-picker bounding box within 390px viewport
```

## Preconditions

- Playwright available; deep `WebWorkingDir` set by parent.
- No expand interaction — assert default render only.

## Steps

1. Start web with open API and deep cwd.
2. Open `/`; assert shortened label, full path in `title`, runner-picker in viewport.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-long-collapsed-default"

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	full := jsString(req.WorkspacePath)
	expectedShort := jsString(shortWorkspaceLabel(req.WorkspacePath))

	body := clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const text = await workspaceLabelText();
const title = (await page.locator('[data-testid="workspace"]').getAttribute('title') || '').trim();
if (!text) throw new Error('expected workspace label text, got empty');
if (!text.startsWith('…/')) {
  throw new Error('expected shortened workspace label with ellipsis, got: ' + text);
}
if (text !== '` + expectedShort + `') {
  throw new Error('expected short label ` + expectedShort + `, got: ' + text);
}
if (title !== '` + full + `') {
  throw new Error('expected title full path, got: ' + title);
}
const empty = page.locator('[data-testid="empty-state"]');
await empty.waitFor({ state: 'visible', timeout: 15000 });
` + assertRunnerPickerWithinViewport()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
