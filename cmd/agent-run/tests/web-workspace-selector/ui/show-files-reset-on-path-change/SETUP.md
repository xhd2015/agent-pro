# Scenario

**Feature**: show-files state collapses when browse path changes (U5)

```
open /workspace at fixture root
  -> expand workspace-show-files (files visible)
  -> enter dir (src) OR parent / quick / recent
  -> showFiles resets: files hidden again; aria-expanded not true
```

## Preconditions

- Fixture from `makeChooserOptimizeFixture` (`src/` has `inner.txt`).
- Product rule: do **not** remember show-files across folders.
- Expect RED until path-change resets `showFiles` local state.

## Steps

1. Expand files; enter `src`; assert files collapsed again.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "show-files-reset-on-path-change"
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
const showFiles = page.locator('[data-testid="workspace-show-files"]');
await showFiles.waitFor({ state: 'visible', timeout: 10000 });
// Expand at root so files are shown.
await showFiles.click();
await page.waitForTimeout(100);
if ((await showFiles.getAttribute('aria-expanded')) !== 'true') {
  throw new Error('expected expanded before path change');
}
const note = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'note.txt' });
await note.first().waitFor({ state: 'visible', timeout: 10000 });
// Enter subdirectory — path change must re-hide files.
const srcDir = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'src' });
await srcDir.first().waitFor({ state: 'visible', timeout: 10000 });
await srcDir.first().click();
await page.waitForTimeout(200);
const afterDir = (await browse.innerText()).trim();
if (!afterDir.endsWith('/src') && !afterDir.endsWith('src')) {
  throw new Error('expected browse path under src, got ' + afterDir);
}
// Files hidden again (inner.txt must not show until re-expand).
const fileCount = await page.locator('[data-testid="workspace-browser-entry"][data-entry-type="file"]').count();
if (fileCount > 0) {
  throw new Error('after path change expected files hidden, file entries=' + fileCount);
}
const inner = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'inner.txt' });
if (await inner.count() > 0 && await inner.first().isVisible().catch(() => false)) {
  throw new Error('inner.txt must be hidden after enter dir until re-expand');
}
// Toggle present (subdir has a file) and collapsed.
const showAfter = page.locator('[data-testid="workspace-show-files"]');
await showAfter.waitFor({ state: 'visible', timeout: 10000 });
const exp = await showAfter.getAttribute('aria-expanded');
if (exp === 'true') {
  throw new Error('after path change show-files must re-hide (aria-expanded!=true), got ' + exp);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
