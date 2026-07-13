# Scenario

**Feature**: Use this folder commits selection and returns home (U5)

```
open /workspace -> browse to SelectPath -> Use this folder
  -> PUT committed
  -> navigate /
  -> home workspace label matches SelectPath
  -> status.workspace == SelectPath
```

## Preconditions

- Two dirs: initial selected vs target SelectPath.
- Expect RED until Use CTA + PUT + home refresh exist.

## Steps

1. Seed initial selected (cwd-ish); create target dir; open selector; Use.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "use-this-folder-commits"
	initial := makeSelectDir(t, req, "initial-ws")
	target := makeSelectDir(t, req, "use-target")
	// Nested under initial so user can enter via browser from initial.
	// Simpler: seed selected=initial, recent includes target, open recent then Use.
	req.SelectPath = target
	req.WebWorkingDir = mustMkdir(t, filepath.Join(req.TempDir, "proc-cwd"))
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": initial,
		"recent_workspaces":  []string{initial, target},
	}); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	targetJS := jsString(target)
	token := jsString(req.Token)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const root = page.locator('[data-testid="workspace-selector"]');
await root.waitFor({ state: 'visible', timeout: 15000 });
const target = ` + targetJS + `;
// Navigate browse path to target via recent row (browse only), then Use.
let row = page.locator('[data-testid="workspace-recent-item"][data-path="' + target + '"]');
if (await row.count() === 0) {
  row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: target });
}
if (await row.count() === 0) {
  const base = target.split('/').filter(Boolean).pop();
  row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: base });
}
await row.first().waitFor({ state: 'visible', timeout: 10000 });
await row.first().click();
await page.waitForTimeout(100);
const useBtn = page.locator('[data-testid="workspace-use-folder"]');
await useBtn.waitFor({ state: 'visible', timeout: 10000 });
await useBtn.click();
await page.waitForURL(url => {
  const p = new URL(url).pathname;
  return p === '/' || p === '';
}, { timeout: 15000 });
// Home shows new workspace
const label = page.locator('[data-testid="workspace-label"], [data-testid="workspace"]');
await label.first().waitFor({ state: 'visible', timeout: 15000 });
const text = (await label.first().innerText()).trim();
const base = target.split('/').filter(Boolean).pop();
if (!text.includes(base) && text !== target && !text.includes('…')) {
  // allow short label containing last segment
  throw new Error('home workspace label should reflect selected path, got ' + text);
}
const status = await page.evaluate(async (tok) => {
  const r = await fetch('/api/agent-run/status', {
    headers: { Authorization: 'Bearer ' + tok }
  });
  return { ok: r.ok, body: await r.json() };
}, ` + token + `);
if (!status.ok) {
  throw new Error('status fetch failed after Use');
}
const ws = (status.body.workspace || '').trim();
if (ws !== target && ws.replace(/\/$/, '') !== target.replace(/\/$/, '')) {
  throw new Error('after Use status.workspace=' + ws + ' want=' + target);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
