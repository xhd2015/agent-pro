# Scenario

**Feature**: dot directories are listed and enterable in the browser (U4)

```
open /workspace at fixture with .hidden-dir/
  -> .hidden-dir visible as dir entry
  -> click enters; browser path ends with .hidden-dir
```

## Preconditions

- Fixture from `makeChooserOptimizeFixture` includes `.hidden-dir/`.
- API must return dot dirs; UI must render them as enterable dirs.
- Expect RED until fs/list includes dots and UI shows them.

## Steps

1. Open selector at fixture; assert `.hidden-dir` visible; enter it.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "browse-dot-dir-enterable"
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
// Dot dir should be listed (dirs are always shown; not behind show-files).
const dotDir = page.locator('[data-testid="workspace-browser-entry"]').filter({ hasText: '.hidden-dir' });
await dotDir.first().waitFor({ state: 'visible', timeout: 10000 });
const entryType = await dotDir.first().getAttribute('data-entry-type');
if (entryType && entryType !== 'dir') {
  throw new Error('.hidden-dir should be type dir, got ' + entryType);
}
await dotDir.first().click();
await page.waitForTimeout(200);
const after = (await browse.innerText()).trim();
if (!after.endsWith('/.hidden-dir') && !after.endsWith('.hidden-dir')) {
  throw new Error('expected browse path under .hidden-dir, got ' + after);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
