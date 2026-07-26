# Scenario

**Feature**: enter a session from home then back preserves session-list scrollTop

```
seed ≥35 sessions → scroll mid-list → open session-item → back
  → scrollTop within ~60px of pre-nav value
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded ≥35 home sessions for mid-list scroll.
- Open API.
- Label: `ui-automation`.

## Steps

1. Seed 40 home sessions; start web.
2. Open `/`; scroll to mid; record scrollTop; click a mid item; back via `.back-link`.
3. Assert restored scrollTop within 60px of pre-nav.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-enter-session-preserves-scroll"
	req.WebTokenMode = "omit"
	runner := "fake-codex"

	if err := seedManyHomeSessions(t, req.Home, runner, 40); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := openHomePage(req.BaseURL) + `
{
  const list = page.locator('[data-testid="session-list"]');
  await list.waitFor({ state: 'visible', timeout: 20000 });
  await page.waitForTimeout(400);
  const itemCount = await page.locator('[data-testid="session-item"]').count();
  if (itemCount < 5) throw new Error('need multiple session-items, got ' + itemCount);

  const scrollTopBeforeNav = await page.evaluate(() => {
    const el = document.querySelector('[data-testid="session-list"]');
    if (!el) return 0;
    const mid = Math.max(120, Math.floor((el.scrollHeight - el.clientHeight) * 0.45));
    el.dispatchEvent(new WheelEvent('wheel', { deltaY: 40, bubbles: true }));
    el.scrollTop = mid;
    el.dispatchEvent(new Event('scroll', { bubbles: true }));
    return el.scrollTop;
  });
  await page.waitForTimeout(250);
  if (scrollTopBeforeNav < 80) {
    throw new Error('could not establish mid scroll before nav: ' + scrollTopBeforeNav);
  }

  const clickIndex = itemCount > 3 ? Math.min(4, itemCount - 1) : 0;
  // Re-apply mid scroll and fire mousedown so persist handlers run (do not scrollIntoView).
  await page.evaluate((idx) => {
    const el = document.querySelector('[data-testid="session-list"]');
    if (!el) return;
    const mid = Math.max(120, Math.floor((el.scrollHeight - el.clientHeight) * 0.45));
    el.scrollTop = mid;
    el.dispatchEvent(new Event('scroll', { bubbles: true }));
    const items = el.querySelectorAll('[data-testid="session-item"]');
    const item = items[idx];
    if (item) {
      item.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
    }
  }, clickIndex);
  await page.waitForTimeout(100);
  await page.locator('[data-testid="session-item"]').nth(clickIndex).click({ force: true });
  await page.waitForURL(/\/sessions\//, { timeout: 15000 });
  await page.waitForTimeout(300);

  const back = page.locator('.back-link').first();
  await back.waitFor({ state: 'visible', timeout: 10000 });
  await back.click();
  await page.waitForURL((url) => {
    try { return new URL(url).pathname === '/'; } catch { return false; }
  }, { timeout: 15000 });
  await page.waitForSelector('[data-testid="session-list"]', { timeout: 15000 });
  // Allow restore layout + rAF timeouts to re-apply scroll.
  await page.waitForTimeout(500);

  const scrollTopAfterBack = await page.evaluate(
    () => document.querySelector('[data-testid="session-list"]')?.scrollTop ?? 0,
  );
  if (Math.abs(scrollTopAfterBack - scrollTopBeforeNav) > 60) {
    throw new Error(
      'session enter/back lost scroll: before=' + scrollTopBeforeNav +
      ' after=' + scrollTopAfterBack,
    );
  }
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
