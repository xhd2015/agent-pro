# Scenario

**Feature**: home → session client nav without full document reload (P1)

```
seed session -> open / with token + __SPA_NAV_MARKER
  -> click [data-testid="session-item"]
  -> URL /sessions/<id>, [data-testid="chat-active"], marker still alive
```

## Preconditions

- Seeded flat session listed on home.
- Token in localStorage so auth gate is skipped.

## Steps

1. Seed `spa-nav-home-sess`.
2. Start web; playwright: marker + token, goto `/`, click session-item, assert URL + chat-active + marker.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "home-to-session-no-reload"
	req.SessionID = "spa-nav-home-sess"
	if err := seedFlatSession(t, req.Home, req.SessionID, "fake-codex", "idle"); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := spaNavMarkerInit() + seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const item = page.locator('[data-testid="session-item"]').first();
await item.waitFor({ state: 'visible', timeout: 15000 });
await item.click();
await page.waitForURL(/\/sessions\//, { timeout: 15000 });
const path = new URL(page.url()).pathname;
const expected = '/sessions/` + req.SessionID + `';
if (path !== expected && decodeURIComponent(path) !== expected) {
  throw new Error('expected path ' + expected + ', got ' + path);
}
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertSpaNavMarkerSurvives()

	req.PlaywrightScript = mobileViewportOptional(body)
	return nil
}
```
