# Scenario

**Feature**: soft auth — token submit without full document reload (P4)

```
explicit --token; empty localStorage -> auth-page
  -> fill token, submit -> home (empty-state or list)
  -> window.__SPA_NAV_MARKER still alive (no location hard reload)
```

## Preconditions

- `WebTokenMode=explicit`, `Token=test-token`.
- Browser starts with **no** `agent-run-token` in localStorage.
- Implementer replaces `window.location.href = '/'` with soft navigate / gate flip.

## Steps

1. Start web with explicit token.
2. Playwright: clear token, set marker, open `/`, submit auth form, assert home + marker.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "soft-auth-no-reload"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := spaNavMarkerInit() + clearTokenInPage() + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const auth = page.locator('[data-testid="auth-page"]');
await auth.waitFor({ state: 'visible', timeout: 15000 });
const input = page.locator('[data-testid="auth-token-input"], [data-testid="auth-page"] input').first();
await input.waitFor({ state: 'visible', timeout: 15000 });
await input.fill('` + req.Token + `');
// Prefer form submit; fall back to Enter.
const form = page.locator('[data-testid="auth-page"] form');
if (await form.count() > 0) {
  await form.evaluate((f) => f.requestSubmit ? f.requestSubmit() : f.submit());
} else {
  await input.press('Enter');
}
const home = page.locator('[data-testid="home-active"], [data-testid="empty-state"], [data-testid="session-list"], [data-testid="composer"]');
await home.first().waitFor({ state: 'visible', timeout: 15000 });
// Auth page must go away after soft auth.
const authStill = await auth.isVisible().catch(() => false);
if (authStill) {
  throw new Error('auth-page still visible after token submit');
}
` + assertSpaNavMarkerSurvives()

	req.PlaywrightScript = mobileViewportOptional(body)
	return nil
}
```
