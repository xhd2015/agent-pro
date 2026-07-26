# Scenario

**Feature**: Cancel returns home without committing selection (U7)

```
selected = A; open selector; browse to B via recent; Cancel
  -> URL /
  -> status.workspace still A
```

## Preconditions

- Seeded selected A; recent includes B.
- Expect RED until Cancel exists.

## Steps

1. Seed A + B; open selector; browse B; Cancel; assert status still A.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "cancel-no-commit"
	a := makeSelectDir(t, req, "cancel-a")
	b := makeSelectDir(t, req, "cancel-b")
	req.SelectPath = a
	req.WebWorkingDir = mustMkdir(t, filepath.Join(req.TempDir, "proc-cwd"))
	if err := writeHomeConfig(t, req.Home, map[string]any{
		"selected_workspace": a,
		"recent_workspaces":  []string{a, b},
	}); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	aJS := jsString(a)
	bJS := jsString(b)
	token := jsString(req.Token)
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/workspace', { waitUntil: 'networkidle' });
const root = page.locator('[data-testid="workspace-selector"]');
await root.waitFor({ state: 'visible', timeout: 15000 });
const bPath = ` + bJS + `;
let row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: bPath });
if (await row.count() === 0) {
  const base = bPath.split('/').filter(Boolean).pop();
  row = page.locator('[data-testid="workspace-recent-item"]').filter({ hasText: base });
}
if (await row.count() > 0) {
  await row.first().click();
  await page.waitForTimeout(100);
}
const cancel = page.locator('[data-testid="workspace-cancel"]');
await cancel.waitFor({ state: 'visible', timeout: 10000 });
await cancel.click();
await page.waitForURL(url => {
  const p = new URL(url).pathname;
  return p === '/' || p === '';
}, { timeout: 15000 });
const status = await page.evaluate(async (tok) => {
  const r = await fetch('/api/agent-run/status', {
    headers: { Authorization: 'Bearer ' + tok }
  });
  return { ok: r.ok, body: await r.json() };
}, ` + token + `);
if (!status.ok) {
  throw new Error('status fetch failed after Cancel');
}
const ws = (status.body.workspace || '').trim();
const want = ` + aJS + `;
if (ws !== want && ws.replace(/\/$/, '') !== want.replace(/\/$/, '')) {
  throw new Error('Cancel must not commit: status.workspace=' + ws + ' want=' + want);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
