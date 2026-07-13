# Scenario

**Feature**: Quick Server cwd chip only sets browse path; does not PUT

```
open /workspace -> tap workspace-quick-cwd
  -> browser path shows process cwd
  -> status.workspace still already-selected
```

## Preconditions

- Config selected_workspace = already-selected (parent browse-only).
- Process started with known `WebWorkingDir` (parent).
- Expect RED until chip exists.

## Steps

1. Start web; open selector; tap Quick cwd; assert path + status.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "quick-cwd-no-commit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	selected := jsString(req.SelectPath)
	cwdJS := jsString(req.WebWorkingDir)
	token := jsString(req.Token)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const root = page.locator('[data-testid="workspace-selector"]');
await root.waitFor({ state: 'visible', timeout: 15000 });
const chip = page.locator('[data-testid="workspace-quick-cwd"]');
await chip.waitFor({ state: 'visible', timeout: 10000 });
await chip.click();
await page.waitForTimeout(150);
const browse = page.locator('[data-testid="workspace-browser-path"]');
await browse.waitFor({ state: 'visible', timeout: 10000 });
const browseText = (await browse.innerText()).trim();
const cwd = ` + cwdJS + `;
if (browseText !== cwd && browseText.replace(/\/$/, '') !== cwd.replace(/\/$/, '')) {
  // allow showing last segments
  const base = cwd.split('/').filter(Boolean).pop();
  if (!browseText.includes(base)) {
    throw new Error('Quick cwd should set browse path to process cwd, got ' + browseText + ' want ' + cwd);
  }
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
  throw new Error('Quick cwd must not commit: status.workspace=' + ws + ' want=' + want);
}
const path = new URL(page.url()).pathname;
if (path !== '/workspace' && path !== '/workspace/') {
  throw new Error('Quick cwd must stay on /workspace, got ' + path);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
