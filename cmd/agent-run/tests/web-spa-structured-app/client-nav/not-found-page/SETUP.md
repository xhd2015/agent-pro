# Scenario

**Feature**: unknown client route shows NotFound page with home control (P3)

```
open /this-route-does-not-exist -> [data-testid="not-found"]
  -> click [data-testid="not-found-home"] -> URL /
```

## Preconditions

- Token seeded so auth page does not mask NotFound.
- Server SPA-fallbacks HTML for unknown path; client router renders NotFound.

## Steps

1. Start web (no session seed required).
2. Playwright: goto unknown path; assert not-found; click home; assert `/`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "not-found-page"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := spaNavMarkerInit() + seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/this-route-does-not-exist', { waitUntil: 'networkidle' });
const nf = page.locator('[data-testid="not-found"]');
await nf.waitFor({ state: 'visible', timeout: 15000 });
const homeCtrl = page.locator('[data-testid="not-found-home"]');
await homeCtrl.waitFor({ state: 'visible', timeout: 15000 });
await homeCtrl.click();
await page.waitForURL((url) => {
  const p = new URL(url).pathname;
  return p === '/' || p === '';
}, { timeout: 15000 });
` + assertSpaNavMarkerSurvives()

	req.PlaywrightScript = mobileViewportOptional(body)
	return nil
}
```
