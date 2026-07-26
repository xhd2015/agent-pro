# Scenario

**Feature**: home workspace control opens `/workspace` selector page (U1)

```
open / -> tap [data-testid="workspace"] (open selector)
  -> URL /workspace
  -> [data-testid="workspace-selector"] visible
```

## Preconditions

- Home renders workspace control.
- Expect RED until control navigates to `/workspace` (today may only expand path).

## Steps

1. Start web; open `/`; click workspace open control; assert URL + selector root.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "open-selector"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const openCtrl = page.locator('[data-testid="workspace-open-selector"], [data-testid="workspace"]');
await openCtrl.first().waitFor({ state: 'visible', timeout: 15000 });
await openCtrl.first().click();
await page.waitForURL(/\/workspace\/?$/, { timeout: 15000 });
const path = new URL(page.url()).pathname;
if (path !== '/workspace' && path !== '/workspace/') {
  throw new Error('expected /workspace, got ' + path);
}
const sel = page.locator('[data-testid="workspace-selector"]');
await sel.waitFor({ state: 'visible', timeout: 15000 });
const useBtn = page.locator('[data-testid="workspace-use-folder"]');
await useBtn.waitFor({ state: 'visible', timeout: 10000 });
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
