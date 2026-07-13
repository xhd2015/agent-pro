# Scenario

**Feature**: show-files toggle sits after last directory entry (layout)

```
open /workspace at fixture with dirs + files
  -> browser list document order (collapsed):
       dir… → workspace-show-files  (no file rows)
  -> expand workspace-show-files
  -> document order (expanded):
       dir… → workspace-show-files → file…
```

## Preconditions

- Fixture from `makeChooserOptimizeFixture` (dirs: `.git`, `.hidden-dir`, `src`;
  files: `.env`, `a.txt`, `note.txt`).
- Seed `selected_workspace` = fixture root so browser opens there.
- Pure layout contract: toggle must appear **after the last directory row** and
  **before any file rows** when expanded. Not only “visible”.
- Expect RED until product moves `workspace-show-files` from above the list into
  that inter-group position.

## Steps

1. Build fixture; seed config; open `/workspace`.
2. Collapsed: collect document order of entry types + show-files; assert
   all `dir` then `show-files`, zero `file`.
3. Expand show-files; re-collect order; assert
   all `dir` then `show-files` then one or more `file`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "show-files-toggle-after-last-dir"
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

	// Collect document-order tokens for browser entries + show-files control.
	// Scope: Browse folders section (aria-label) so path/parent are excluded.
	const collectOrderJS = `
async function collectBrowserOrder(page) {
  return page.evaluate(() => {
    const section =
      document.querySelector('[aria-label="Browse folders"]') ||
      document.querySelector('.workspace-browser') ||
      document.querySelector('[data-testid="workspace-selector"]');
    if (!section) return [];
    const nodes = section.querySelectorAll(
      '[data-testid="workspace-browser-entry"], [data-testid="workspace-show-files"]'
    );
    return Array.from(nodes).map((el) => {
      const tid = el.getAttribute('data-testid') || '';
      if (tid === 'workspace-show-files') return 'show-files';
      const typ = el.getAttribute('data-entry-type');
      if (typ === 'dir' || typ === 'file') return typ;
      return 'entry';
    });
  });
}
function assertOrderCollapsed(order) {
  if (!Array.isArray(order) || order.length === 0) {
    throw new Error('collapsed order empty: ' + JSON.stringify(order));
  }
  if (order.includes('file')) {
    throw new Error('collapsed must have no file entries, order=' + JSON.stringify(order));
  }
  const showIdx = order.indexOf('show-files');
  if (showIdx < 0) {
    throw new Error('collapsed missing show-files, order=' + JSON.stringify(order));
  }
  const dirs = order.slice(0, showIdx);
  const after = order.slice(showIdx + 1);
  if (dirs.length === 0 || dirs.some((x) => x !== 'dir')) {
    throw new Error(
      'collapsed: all items before show-files must be dir, order=' + JSON.stringify(order)
    );
  }
  if (after.length !== 0) {
    throw new Error(
      'collapsed: nothing after show-files allowed, order=' + JSON.stringify(order)
    );
  }
  // Toggle must not sit before the first dir (current product places it above the list).
  if (showIdx === 0) {
    throw new Error(
      'collapsed: show-files is before any dir (want after last dir), order=' +
        JSON.stringify(order)
    );
  }
}
function assertOrderExpanded(order) {
  if (!Array.isArray(order) || order.length === 0) {
    throw new Error('expanded order empty: ' + JSON.stringify(order));
  }
  const showIdx = order.indexOf('show-files');
  if (showIdx < 0) {
    throw new Error('expanded missing show-files, order=' + JSON.stringify(order));
  }
  const before = order.slice(0, showIdx);
  const after = order.slice(showIdx + 1);
  if (before.length === 0 || before.some((x) => x !== 'dir')) {
    throw new Error(
      'expanded: all items before show-files must be dir, order=' + JSON.stringify(order)
    );
  }
  if (after.length === 0 || after.some((x) => x !== 'file')) {
    throw new Error(
      'expanded: all items after show-files must be file (at least one), order=' +
        JSON.stringify(order)
    );
  }
  if (showIdx === 0) {
    throw new Error(
      'expanded: show-files is before any dir (want after last dir), order=' +
        JSON.stringify(order)
    );
  }
}
`

	body := seedTokenInPage(req.Token) + collectOrderJS + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const rootEl = page.locator('[data-testid="workspace-selector"]');
await rootEl.waitFor({ state: 'visible', timeout: 15000 });
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
const anyDir = page.locator('[data-testid="workspace-browser-entry"][data-entry-type="dir"]');
await anyDir.first().waitFor({ state: 'visible', timeout: 10000 });
const showFiles = page.locator('[data-testid="workspace-show-files"]');
await showFiles.waitFor({ state: 'visible', timeout: 10000 });
const expanded0 = await showFiles.getAttribute('aria-expanded');
if (expanded0 === 'true') {
  throw new Error('show-files should start collapsed, aria-expanded=' + expanded0);
}
// L1 collapsed: dirs… → show-files (no files)
const collapsedOrder = await collectBrowserOrder(page);
assertOrderCollapsed(collapsedOrder);
// Expand
await showFiles.click();
await page.waitForTimeout(100);
if ((await showFiles.getAttribute('aria-expanded')) !== 'true') {
  throw new Error('after expand aria-expanded should be true');
}
const note = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'note.txt' });
await note.first().waitFor({ state: 'visible', timeout: 10000 });
// L2 expanded: dirs… → show-files → files…
const expandedOrder = await collectBrowserOrder(page);
assertOrderExpanded(expandedOrder);
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
