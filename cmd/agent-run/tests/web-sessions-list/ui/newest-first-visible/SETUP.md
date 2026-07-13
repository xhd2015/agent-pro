# Scenario

**Feature**: home session list shows newest session first (UI)

```
seed 5 controlled updated_at
open /
  -> first [data-testid="session-item"] preview contains "brand newest epsilon"
  -> first item is not the oldest ("hello world alpha")
```

## Preconditions

- `defaultFiveSessions()`; newest prompt = `brand newest epsilon`.
- Expect RED while UI uses `sortSessionsOldestFirst`.

## Steps

1. Seed five flat sessions; start web.
2. Open home; wait for session-list; assert first row is newest.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "newest-first-visible"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + openHomeWithSessions(req.BaseURL) + `
const items = page.locator('[data-testid="session-item"]');
await items.first().waitFor({ state: 'visible', timeout: 15000 });
const count = await items.count();
if (count < 5) throw new Error('expected >=5 session items, got ' + count);
const firstPreview = page.locator('[data-testid="session-item"]').first().locator('[data-testid="session-preview"]');
const text = (await firstPreview.innerText()).trim();
if (!/brand newest epsilon/i.test(text)) {
  throw new Error('first row should be newest (epsilon), got: ' + text);
}
if (/hello world alpha/i.test(text)) {
  throw new Error('first row is oldest (alpha) — still oldest-first: ' + text);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
