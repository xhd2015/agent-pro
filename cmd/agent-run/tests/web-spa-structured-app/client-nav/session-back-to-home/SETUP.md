# Scenario

**Feature**: session page → home via back link (P2)

```
seed + open /sessions/:id -> click .back-link -> URL / + home UI (home-active or empty-state or session-list)
```

## Preconditions

- Seeded session so session detail can load after token seed.
- Prefer `.back-link` (current “← Sessions” control).

## Steps

1. Seed session; start web.
2. Playwright: open session path with token; click `.back-link`; assert `/` and home UI.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "session-back-to-home"
	req.SessionID = "spa-nav-back-sess"
	if err := seedFlatSession(t, req.Home, req.SessionID, "fake-codex", "idle"); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := spaNavMarkerInit() + seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + `/sessions/` + req.SessionID + `', { waitUntil: 'networkidle' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
const back = page.locator('.back-link, [data-testid="back-to-sessions"], a[href="/"]').first();
await back.waitFor({ state: 'visible', timeout: 15000 });
await back.click();
await page.waitForURL((url) => {
  const p = new URL(url).pathname;
  return p === '/' || p === '';
}, { timeout: 15000 });
const home = page.locator('[data-testid="home-active"], [data-testid="empty-state"], [data-testid="session-list"]');
await home.first().waitFor({ state: 'visible', timeout: 15000 });
` + assertSpaNavMarkerSurvives()

	req.PlaywrightScript = mobileViewportOptional(body)
	return nil
}
```
