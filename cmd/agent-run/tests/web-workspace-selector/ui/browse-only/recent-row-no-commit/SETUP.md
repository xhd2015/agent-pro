# Scenario

**Feature**: Recent row only sets browse path; does not PUT (U4)

```
seed recent_workspaces = [selected, other]
open /workspace -> tap recent item for other
  -> browser path == other
  -> status.workspace still selected
```

## Preconditions

- Config has two recent entries; selected is first path.
- Expect RED until Recent UI + browse-only behavior exist.

## Steps

1. Seed config with two recents; start web; tap recent; assert no commit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "recent-row-no-commit"
	// Parent already set SelectPath as selected; add a second recent target.
	other := makeSelectDir(t, req, "recent-other")
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": req.SelectPath,
		"recent_workspaces":  []string{req.SelectPath, other},
	}); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	selected := jsString(req.SelectPath)
	otherJS := jsString(other)
	token := jsString(req.Token)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const root = page.locator('[data-testid="workspace-selector"]');
await root.waitFor({ state: 'visible', timeout: 15000 });
const otherPath = ` + otherJS + `;
// Prefer data-path attribute; fall back to text match.
let row = page.locator('[data-testid="workspace-recent-item"][data-path="' + otherPath + '"]');
if (await row.count() === 0) {
  row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: otherPath });
}
if (await row.count() === 0) {
  // last segment match
  const base = otherPath.split('/').filter(Boolean).pop();
  row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: base });
}
await row.first().waitFor({ state: 'visible', timeout: 10000 });
await row.first().click();
await page.waitForTimeout(150);
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
const browseText = (await browse.innerText()).trim();
if (browseText !== otherPath && !browseText.includes(otherPath.split('/').filter(Boolean).pop())) {
  throw new Error('recent tap should set browse path to other, got ' + browseText);
}
const status = await page.evaluate(async (tok) => {
  const r = await fetch('/api/agent-run/status', {
    headers: { Authorization: 'Bearer ' + tok }
  });
  return { ok: r.ok, status: r.status, body: await r.json() };
}, ` + token + `);
if (!status.ok) {
  throw new Error('status fetch failed: ' + status.status);
}
const ws = (status.body.workspace || '').trim();
const want = ` + selected + `;
if (ws !== want && ws.replace(/\/$/, '') !== want.replace(/\/$/, '')) {
  throw new Error('Recent must not commit: status.workspace=' + ws + ' want=' + want);
}
const path = new URL(page.url()).pathname;
if (path !== '/workspace' && path !== '/workspace/') {
  throw new Error('Recent must stay on /workspace, got ' + path);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
