# Scenario

**Feature**: home loads page size 30 then **explicit** load-more button appends

```
seed 35 sessions (page-seed-001 … page-seed-035)
open /
  -> initial session-item count <= 30
  -> click [data-testid="session-load-more"] (button only — no scroll fallback)
  -> item count increases beyond initial (still <= 35)
```

## Preconditions

- Product page size **30**; 35 seeds so has_more is true after first page.
- Must use explicit `session-load-more` button click — scroll near bottom must
  **not** be used as a load trigger in this leaf.
- Label: `ui-automation`.

## Steps

1. Seed 35 flat sessions; start web.
2. Open home; measure initial count; click load-more; assert growth.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "load-more-appends"
	const n = 35
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
const loadMore = page.locator('[data-testid="session-load-more"]');
await loadMore.first().waitFor({ state: 'visible', timeout: 15000 });
// Button-only: do not scroll-as-load. Click the explicit control.
await loadMore.first().click({ force: true });
await page.waitForTimeout(1000);
const after = await page.locator('[data-testid="session-item"]').count();
if (after <= initial) {
  throw new Error('load more button should append: initial=' + initial + ' after=' + after);
}
if (after > %d) {
  throw new Error('after load-more count ' + after + ' exceeds seed total %d');
}
`, pageSize, pageSize, n, n)
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
