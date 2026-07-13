# Scenario

**Feature**: copy control writes full absolute workspace path to clipboard

```
# expand then copy
open / (deep) -> expand WorkspacePath -> click [data-testid="workspace-copy"]
  -> navigator.clipboard.readText() === full absolute path
```

## Preconditions

- Clipboard permissions granted to the origin via Playwright context.
- Copy control required by option A (expect RED until implemented).

## Steps

1. Start web; open `/`; grant clipboard-read/write.
2. Expand path; click `workspace-copy`.
3. Read clipboard; must equal full workspace path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-long-copy-full-path"

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	full := jsString(req.WorkspacePath)

	body := clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const ctx = page.context();
await ctx.grantPermissions(['clipboard-read', 'clipboard-write'], { origin: '` + req.BaseURL + `' });
` + jsWaitWorkspaceVisible() + jsWorkspaceLabelText() + `
const toggle = page.locator('[data-testid="workspace-toggle"]');
await toggle.waitFor({ state: 'visible', timeout: 15000 });
await toggle.click();
await page.waitForTimeout(100);
const expandedText = await workspaceLabelText();
if (expandedText !== '` + full + `') {
  throw new Error('must expand before copy; label=' + expandedText);
}
const copyBtn = page.locator('[data-testid="workspace-copy"]');
await copyBtn.waitFor({ state: 'visible', timeout: 15000 });
await copyBtn.click();
await page.waitForTimeout(100);
const clip = await page.evaluate(async () => navigator.clipboard.readText());
if (clip !== '` + full + `') {
  throw new Error('clipboard expected full path, got: ' + JSON.stringify(clip));
}
`

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
