# Scenario

**Feature**: Quick Home chip only sets browse path; does not PUT (U3)

```
open /workspace -> tap workspace-quick-home
  -> browser path shows OS home
  -> status.workspace still already-selected
```

## Preconditions

- Config selected_workspace = `already-selected` dir (parent).
- Expect RED until selector + browse-only chips exist.

## Steps

1. Start web; open selector; tap Quick Home; assert path + status via fetch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "quick-home-no-commit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	selected := jsString(req.SelectPath)
	token := jsString(req.Token)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const root = page.locator('[data-testid="workspace-selector"]');
await root.waitFor({ state: 'visible', timeout: 15000 });
const chip = page.locator('[data-testid="workspace-quick-home"]');
await chip.waitFor({ state: 'visible', timeout: 10000 });
await chip.click();
await page.waitForTimeout(150);
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
const browseText = (await browse.innerText()).trim();
// Browse path should reflect user home (or at least change toward home).
if (!browseText || browseText.length < 1) {
  throw new Error('browser path empty after Quick Home');
}
// Critical: selected workspace via status must remain unchanged (no auto PUT).
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
if (ws !== ` + selected + ` && ws.replace(/\/$/, '') !== ` + selected + `.replace(/\/$/, '')) {
  throw new Error('Quick Home must not commit: status.workspace=' + ws + ' want=' + ` + selected + `);
}
// Stay on selector (no auto navigate home).
const path = new URL(page.url()).pathname;
if (path !== '/workspace' && path !== '/workspace/') {
  throw new Error('Quick Home must stay on /workspace, got ' + path);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
