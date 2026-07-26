# Scenario

**Feature**: home composer draft survives open selector + Cancel (U2)

```
type composer draft -> open /workspace -> Cancel
  -> back on /
  -> composer-input value still equals draft text
```

## Preconditions

- Draft owned at App level (not local HomePage-only state).
- Today HomePage `useState` draft is lost on unmount → expect RED.

## Steps

1. Start web; type draft; open selector; cancel; assert draft.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "draft-survives-roundtrip"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	const draft = "selector draft must survive"
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const input = page.locator('[data-testid="composer-input"]');
await input.waitFor({ state: 'visible', timeout: 15000 });
await input.fill(` + jsString(draft) + `);
const openCtrl = page.locator('[data-testid="workspace-open-selector"], [data-testid="workspace"]');
await openCtrl.first().click();
await page.waitForURL(/\/workspace\/?$/, { timeout: 15000 });
const cancel = page.locator('[data-testid="workspace-cancel"]');
await cancel.waitFor({ state: 'visible', timeout: 10000 });
await cancel.click();
await page.waitForURL(/\/?$/, { timeout: 15000 });
// Allow home to remount
await page.waitForTimeout(200);
const again = page.locator('[data-testid="composer-input"]');
await again.waitFor({ state: 'visible', timeout: 15000 });
const val = await again.inputValue();
if (val !== ` + jsString(draft) + `) {
  throw new Error('draft lost after selector cancel: got ' + JSON.stringify(val));
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
