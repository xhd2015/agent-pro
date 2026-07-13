# Scenario

**Feature**: browse enters directories; files shown non-selectable after expand (U6)

```
open /workspace with browse path = fixture root
  -> dirs visible; files hidden by default
  -> expand workspace-show-files
  -> note.txt visible, disabled / non-navigable
  -> click subdir -> browser path ends with /subdir
```

## Preconditions

- Fixture tree: `subdir/` + `note.txt`.
- Seed selected + open browser at fixture root.
- **Update for dir-chooser optimize**: files are no longer visible by default —
  expand `workspace-show-files` before asserting file disabled.

## Steps

1. Build fixture; seed selected=fixture; open selector; expand files; assert disabled; enter dir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "browse-enter-dir-files-disabled"
	root := makeFixtureTree(t, req)
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

	rootJS := jsString(root)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const rootEl = page.locator('[data-testid="workspace-selector"]');
await rootEl.waitFor({ state: 'visible', timeout: 15000 });
// Ensure browse starts at fixture root (selected/recent).
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
// Directory entry always visible
const dirEntry = page.locator('[data-testid="workspace-browser-entry"][data-entry-type="dir"], [data-testid="workspace-browser-entry"]').filter({ hasText: 'subdir' });
await dirEntry.first().waitFor({ state: 'visible', timeout: 10000 });
// Expand files first (default hide) — required after dir-chooser optimize
const showFiles = page.locator('[data-testid="workspace-show-files"]');
await showFiles.waitFor({ state: 'visible', timeout: 10000 });
const expanded = await showFiles.getAttribute('aria-expanded');
if (expanded !== 'true') {
  await showFiles.click();
  await page.waitForTimeout(100);
}
// File entry visible after expand
const fileEntry = page.locator('[data-testid="workspace-browser-entry"][data-entry-type="file"], [data-testid="workspace-browser-entry"]').filter({ hasText: 'note.txt' });
await fileEntry.first().waitFor({ state: 'visible', timeout: 10000 });
// File should be disabled / aria-disabled / not navigable
const fileHandle = fileEntry.first();
const disabled = await fileHandle.getAttribute('aria-disabled');
const cls = (await fileHandle.getAttribute('class')) || '';
const pointer = await fileHandle.evaluate(el => getComputedStyle(el).pointerEvents);
if (disabled !== 'true' && !cls.includes('disabled') && pointer !== 'none') {
  // Still try click — path must not become the file path
  const before = (await browse.innerText()).trim();
  await fileHandle.click({ force: true }).catch(() => {});
  await page.waitForTimeout(100);
  const afterFile = (await browse.innerText()).trim();
  if (afterFile.includes('note.txt') && afterFile !== before) {
    throw new Error('file entry must not become browse path: ' + afterFile);
  }
}
// Enter directory
await dirEntry.first().click();
await page.waitForTimeout(150);
const afterDir = (await browse.innerText()).trim();
const rootPath = ` + rootJS + `;
if (!afterDir.endsWith('/subdir') && !afterDir.endsWith('subdir') && afterDir !== rootPath + '/subdir') {
  throw new Error('expected browse path under subdir, got ' + afterDir);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
