# Scenario

**Feature**: multi-step home list scroll settles at C and does not snap back to B after idle/poll

```
seed ≥35 home sessions → / → scroll fractions A→B→C with settle gaps
  → wait ≥3.5s → scrollTop stays near C (not snap to B)
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded ≥35 home sessions so `session-list` overflows and mid scroll has range.
- Open API (`WebTokenMode=omit`).
- Label: `ui-automation, slow` (multi-step + settle wait).

## Steps

1. Seed 40 home sessions; start web open API.
2. Open `/`; wait for overflow list.
3. Scroll to ~25%, ~45%, ~70% with settle gaps; wait ≥3.5s after C.
4. Assert final `scrollTop` stays near C (not within 40px of B while far from C).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-scroll-no-snap-back"
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
  const overflow = await list.evaluate((el) => el.scrollHeight > el.clientHeight + 2);
  if (!overflow) throw new Error('session-list must overflow for multi-step scroll');

  async function scrollListToFraction(frac) {
    return page.evaluate((f) => {
      const el = document.querySelector('[data-testid="session-list"]');
      if (!el) return 0;
      const max = Math.max(0, el.scrollHeight - el.clientHeight);
      const target = Math.floor(max * f);
      el.dispatchEvent(new WheelEvent('wheel', {
        deltaY: target > el.scrollTop ? 40 : -40,
        bubbles: true,
      }));
      el.scrollTop = target;
      el.dispatchEvent(new Event('scroll', { bubbles: true }));
      return el.scrollTop;
    }, frac);
  }

  const stepA = await scrollListToFraction(0.25);
  await page.waitForTimeout(400);
  const stepB = await scrollListToFraction(0.45);
  await page.waitForTimeout(400);
  const stepBSettle = await list.evaluate((el) => el.scrollTop);
  const stepC = await scrollListToFraction(0.7);
  // Wait past poll interval + margin so content refresh would re-pin if buggy.
  await page.waitForTimeout(3500);
  const after = await list.evaluate((el) => el.scrollTop);

  if (stepC < 80) {
    throw new Error('step C scrollTop too small to detect snap-back: ' + stepC);
  }
  const nearC = Math.abs(after - stepC) <= 80;
  const snappedToB =
    stepBSettle >= 80 &&
    Math.abs(after - stepBSettle) < 40 &&
    Math.abs(after - stepC) > 80;
  if (!nearC || snappedToB) {
    throw new Error(
      'multi-step scroll snapped away from C: A=' + stepA +
      ' B=' + stepB + ' B_settle=' + stepBSettle +
      ' C=' + stepC + ' after=' + after,
    );
  }
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
