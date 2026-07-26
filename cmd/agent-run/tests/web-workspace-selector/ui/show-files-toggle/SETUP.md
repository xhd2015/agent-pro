# Scenario

**Feature**: expand/collapse files via workspace-show-files (U2, U3)

```
open /workspace at fixture with files (default hidden)
  -> tap [data-testid="workspace-show-files"]
  -> files appear (note.txt, .env, a.txt); still non-selectable
  -> aria-expanded=true
  -> tap again (collapse)
  -> files hidden; aria-expanded=false
```

## Preconditions

- Fixture from `makeChooserOptimizeFixture`.
- Expect RED until show-files toggle + hide/show files UI exists.

## Steps

1. Open selector; expand; assert files + disabled; collapse; assert hidden.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "show-files-toggle"
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
// Expand
await showFiles.click();
await page.waitForTimeout(100);
const expanded = await showFiles.getAttribute('aria-expanded');
if (expanded !== 'true') {
  throw new Error('after expand aria-expanded should be true, got ' + expanded);
}
// Files visible including regular + dot file
const note = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: 'note.txt' });
await note.first().waitFor({ state: 'visible', timeout: 10000 });
const envEntry = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: '.env' });
await envEntry.first().waitFor({ state: 'visible', timeout: 10000 });
// Files non-selectable / disabled
const fileHandle = note.first();
const disabled = await fileHandle.getAttribute('aria-disabled');
const isDisabled = await fileHandle.isDisabled().catch(() => false);
const cls = (await fileHandle.getAttribute('class')) || '';
if (disabled !== 'true' && !isDisabled && !cls.includes('disabled')) {
  const before = (await browse.innerText()).trim();
  await fileHandle.click({ force: true }).catch(() => {});
  await page.waitForTimeout(100);
  const afterFile = (await browse.innerText()).trim();
  if (afterFile.includes('note.txt') && afterFile !== before) {
    throw new Error('file entry must not become browse path: ' + afterFile);
  }
}
// Collapse
await showFiles.click();
await page.waitForTimeout(100);
const collapsed = await showFiles.getAttribute('aria-expanded');
if (collapsed === 'true') {
  throw new Error('after collapse aria-expanded should not be true, got ' + collapsed);
}
const fileCount = await page.locator('[data-testid="workspace-browser-entry"][data-entry-type="file"]').count();
if (fileCount > 0) {
  throw new Error('after collapse expected 0 file entries, got ' + fileCount);
}
if (await note.count() > 0 && await note.first().isVisible().catch(() => false)) {
  throw new Error('note.txt still visible after collapse');
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
