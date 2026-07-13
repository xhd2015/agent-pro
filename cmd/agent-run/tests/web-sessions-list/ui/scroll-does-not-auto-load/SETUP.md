# Scenario

**Feature**: scrolling near the bottom of the session list does **not** auto-load the next page

```
seed ≥55 sessions; first page 30
open /
  -> initial session-item count === 30 (or ≤30)
  -> scroll session-list to bottom without clicking load-more
  -> wait ≥800ms
  -> item count stays at initial (no infinite scroll)
```

## Preconditions

- Product page size **30**; ≥55 seeds so has_more remains true after first page.
- Explicit load-more is the only pagination trigger.
- Label: `ui-automation`.

## Steps

1. Seed 55 flat sessions; start web.
2. Open home; record item count; scroll list to bottom; assert count stable ≥800ms.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "scroll-does-not-auto-load"
	const n = 55
	if err := seedSessions(t, req, manySessions(n)); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + openHomeWithSessions(req.BaseURL) + fmt.Sprintf(`
const items = page.locator('[data-testid="session-item"]');
await items.first().waitFor({ state: 'visible', timeout: 20000 });
await page.waitForTimeout(500);
const initial = await items.count();
if (initial > %d) {
  throw new Error('initial page should be <= page size %d, got ' + initial);
}
if (initial < 1) {
  throw new Error('expected at least one session-item');
}
// Ensure load-more exists (has_more) so auto-load would have something to fetch.
const loadMoreCount = await page.locator('[data-testid="session-load-more"]').count();
if (loadMoreCount < 1) {
  throw new Error('expected session-load-more when has_more (seed=%d)');
}
const offsetFetches = [];
page.on('request', (req) => {
  if (req.method() !== 'GET') return;
  const u = req.url();
  if (u.includes('/api/agent-run/sessions') && /[?&]offset=([1-9]\d*)/.test(u)) {
    offsetFetches.push(u);
  }
});
// Scroll to bottom WITHOUT clicking load-more.
await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (!list) throw new Error('no session-list');
  list.scrollTop = list.scrollHeight;
  list.dispatchEvent(new Event('scroll', { bubbles: true }));
});
await page.waitForTimeout(900);
const after = await page.locator('[data-testid="session-item"]').count();
if (after > initial) {
  throw new Error(
    'scroll near bottom must not auto-load: initial=' + initial + ' after=' + after +
    ' offsetFetches=' + JSON.stringify(offsetFetches),
  );
}
if (offsetFetches.length > 0) {
  throw new Error('unexpected offset>0 sessions fetch on scroll: ' + JSON.stringify(offsetFetches));
}
`, pageSize, pageSize, n)
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
