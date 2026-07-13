# Scenario

**Feature**: hide show-files control when the listing has zero files

```
open /workspace at dirs-only fixture (alpha/, beta/ only)
  -> dir entries visible
  -> [data-testid="workspace-show-files"] absent / not visible
```

## Preconditions

- Fixture from `makeDirsOnlyFixture` (no files).
- Expect RED until UI omits toggle when `entries` has no files.

## Steps

1. Open selector at dirs-only root; assert no show-files control.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "show-files-control-hidden-when-no-files"
	root := makeDirsOnlyFixture(t, req)
	req.SelectPath = root
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": root,
		"recent_workspaces":  []string{root},
	}); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const rootEl = page.locator('[data-testid="workspace-selector"]');
await rootEl.waitFor({ state: 'visible', timeout: 15000 });
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
const alpha = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'alpha' });
await alpha.first().waitFor({ state: 'visible', timeout: 10000 });
const showFiles = page.locator('[data-testid="workspace-show-files"]');
const count = await showFiles.count();
if (count > 0) {
  const visible = await showFiles.first().isVisible().catch(() => false);
  if (visible) {
    throw new Error('workspace-show-files must be hidden/omitted when zero files in listing');
  }
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
