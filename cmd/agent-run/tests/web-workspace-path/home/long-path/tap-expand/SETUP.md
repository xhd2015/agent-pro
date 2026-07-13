# Scenario

**Feature**: tap workspace control expands to full path on home (option A)

```
# collapsed -> tap toggle -> expanded full path
open / (deep cwd) -> tap [data-testid="workspace-toggle"]
  -> workspace-label === full absolute path
  -> aria-expanded=true; no horizontal document scroll
```

## Preconditions

- Feature not yet implemented: expects RED until `WorkspacePath` toggle exists.
- Deep `WebWorkingDir` from parent; open API home.

## Steps

1. Start web; open `/`; wait for workspace.
2. Click `[data-testid="workspace-toggle"]`.
3. Assert label text equals full path, `aria-expanded=true`, no h-scroll (script end).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-long-tap-expand"

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	full := jsString(req.WorkspacePath)

	body := clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const before = await workspaceLabelText();
if (!before.startsWith('…/')) {
  throw new Error('precondition: expected collapsed short label, got: ' + before);
}
const toggle = page.locator('[data-testid="workspace-toggle"]');
await toggle.waitFor({ state: 'visible', timeout: 15000 });
await toggle.click();
await page.waitForTimeout(100);
const after = await workspaceLabelText();
if (after !== '` + full + `') {
  throw new Error('expected expanded full path, got: ' + after);
}
const expanded = await toggle.getAttribute('aria-expanded');
if (expanded !== 'true') {
  throw new Error('expected aria-expanded=true after expand, got: ' + expanded);
}
const label = page.locator('[data-testid="workspace-label"]');
await label.waitFor({ state: 'visible', timeout: 5000 });
`

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
