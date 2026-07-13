# Scenario

**Feature**: load-more control sits at the end of the scrollable session list

```
seed ≥35 sessions
open /
  -> [data-testid="session-load-more"] is contained by [data-testid="session-list"]
  -> not a sticky sibling always visible at scrollTop≈0 outside the list
```

## Preconditions

- Product page size **30**; ≥35 seeds so has_more is true and button renders.
- Load-more is rendered as `SessionList` footer (inside the scroll nav).
- Label: `ui-automation`.

## Steps

1. Seed 35 flat sessions; start web.
2. Open home; assert load-more is inside session-list DOM and not sticky outside.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "load-more-inside-list"
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
await page.waitForTimeout(400);
const loadMore = page.locator('[data-testid="session-load-more"]');
await loadMore.first().waitFor({ state: 'attached', timeout: 15000 });
const placement = await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  const btn = document.querySelector('[data-testid="session-load-more"]');
  if (!list) return { ok: false, reason: 'no session-list' };
  if (!btn) return { ok: false, reason: 'no session-load-more' };
  const btnInsideList = list.contains(btn);
  const r = btn.getBoundingClientRect();
  const btnVisibleInViewport =
    r.height > 0 && r.bottom > 0 && r.top < window.innerHeight && r.width > 0;
  return {
    ok: true,
    btnInsideList,
    listScrollTop: list.scrollTop,
    listOverflow: list.scrollHeight > list.clientHeight + 2,
    btnVisibleInViewport,
    itemCount: document.querySelectorAll('[data-testid="session-item"]').length,
  };
});
if (!placement.ok) {
  throw new Error('placement probe failed: ' + placement.reason);
}
if (placement.itemCount > %d) {
  throw new Error('initial page should be <= %d, got ' + placement.itemCount);
}
if (!placement.btnInsideList) {
  throw new Error(
    'session-load-more must be contained by session-list (list-end footer), placement=' +
      JSON.stringify(placement),
  );
}
// Sticky sibling regression: button outside list always in viewport at scrollTop=0.
if (
  !placement.btnInsideList &&
  placement.btnVisibleInViewport &&
  (placement.listScrollTop || 0) < 8 &&
  placement.listOverflow
) {
  throw new Error('load-more sticky outside list: ' + JSON.stringify(placement));
}
`, pageSize, pageSize)
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
