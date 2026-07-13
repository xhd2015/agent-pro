# Scenario

**Feature**: home search box filters the visible session list

```
seed 5 (delta prompt has UNIQUE-QUERY-TOKEN)
open /
type UNIQUE-QUERY-TOKEN into [data-testid="session-search"]
  -> only matching session-item(s) remain
  -> preview contains UNIQUE-QUERY-TOKEN / delta
```

## Preconditions

- Search control `data-testid="session-search"` (new UI).
- Expect RED until search input + client/server filter exist.

## Steps

1. Seed five; start web.
2. Open home; fill search; assert filtered list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)
	req.Scenario = "search-filters-list"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + openHomeWithSessions(req.BaseURL) + `
const search = page.locator('[data-testid="session-search"]');
await search.waitFor({ state: 'visible', timeout: 15000 });
await search.fill('UNIQUE-QUERY-TOKEN');
// allow debounce / refetch
await page.waitForTimeout(800);
const items = page.locator('[data-testid="session-item"]');
const count = await items.count();
if (count !== 1) {
  throw new Error('expected 1 filtered session-item, got ' + count);
}
const preview = (await items.first().locator('[data-testid="session-preview"]').innerText()).trim();
if (!/UNIQUE-QUERY-TOKEN/i.test(preview) && !/delta/i.test(preview)) {
  throw new Error('filtered preview should match delta/token, got: ' + preview);
}
// clearing search restores more items
await search.fill('');
await page.waitForTimeout(800);
const restored = await page.locator('[data-testid="session-item"]').count();
if (restored < 2) {
  throw new Error('after clear search expected >=2 items, got ' + restored);
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
