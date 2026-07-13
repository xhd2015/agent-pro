# Scenario

**Feature**: workspace browser hides files by default (U1)

```
open /workspace browse path = fixture root (has dirs + files)
  -> dir entries visible (e.g. src, .hidden-dir)
  -> file entries NOT visible (note.txt, .env, a.txt)
  -> [data-testid="workspace-show-files"] present (files exist)
```

## Preconditions

- Fixture from `makeChooserOptimizeFixture`.
- Seed selected_workspace = fixture root so browser opens there.
- Expect RED until UI defaults to hide files + show-files control.

## Steps

1. Build fixture; seed config; open `/workspace`; assert dirs only + toggle present.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "files-hidden-by-default"
	root := makeChooserOptimizeFixture(t, req)
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
// Wait for at least one dir entry to render.
const anyDir = page.locator('[data-testid="workspace-browser-entry"][data-entry-type="dir"]');
await anyDir.first().waitFor({ state: 'visible', timeout: 10000 });
// Explicit dir names
const srcDir = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'src' });
await srcDir.first().waitFor({ state: 'visible', timeout: 10000 });
// Files must NOT be visible by default.
const fileEntries = page.locator('[data-testid="workspace-browser-entry"][data-entry-type="file"]');
const fileCount = await fileEntries.count();
if (fileCount > 0) {
  throw new Error('files must be hidden by default, found ' + fileCount + ' file entries');
}
const noteVisible = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'note.txt' });
if (await noteVisible.count() > 0 && await noteVisible.first().isVisible().catch(() => false)) {
  throw new Error('note.txt must not be visible by default');
}
// Show-files control required when listing has files.
const showFiles = page.locator('[data-testid="workspace-show-files"]');
await showFiles.waitFor({ state: 'visible', timeout: 10000 });
const expanded = await showFiles.getAttribute('aria-expanded');
if (expanded === 'true') {
  throw new Error('workspace-show-files should start collapsed (aria-expanded!=true), got ' + expanded);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
