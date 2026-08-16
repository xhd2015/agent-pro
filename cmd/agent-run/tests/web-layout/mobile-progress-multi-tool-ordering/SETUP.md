# Scenario

**Feature**: progress compaction preserves chronological order when different tool_call_ids interleave

```
seed tool A -> tool B -> tool A update -> session page -> first progress card is A, second is B
```

## Preconditions

- `playwright-debug` on PATH (`label: chromium`).
- Seeded events include two distinct `tool_call_id` values with a late update to the first tool.
- Frontend compacts duplicate tool rows without moving an earlier tool after a later-started tool.

## Steps

1. Seed `fake-opencode/layout-multi-tool-order` with interleaved tool calls.
2. Start `agent-run web` with explicit token.
3. Open session URL; assert progress card order matches start order.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "progress-multi-tool-ordering"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-multi-tool-order"
	workspacePath := filepath.Join(req.TempDir, "multi-tool-order")

	if err := seedProgressMultiToolOrderingSession(t, req.Home, runner, sessionID, workspacePath); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const progress = page.locator('[data-testid="progress-card"]');
await progress.first().waitFor({ state: 'visible', timeout: 15000 });
const count = await progress.count();
if (count !== 2) throw new Error('expected exactly 2 compacted progress cards, got ' + count);
const labels = await progress.locator('.progress-card-label').allTextContents();
if (!labels.every(t => /tool/i.test(t))) {
  throw new Error('expected only Tool progress labels, got ' + JSON.stringify(labels));
}
const bodies = await progress.locator('.progress-card-body').allTextContents();
if (!/alpha done/i.test(bodies[0] || '')) {
  throw new Error('expected merged alpha output on first tool card, got ' + JSON.stringify(bodies));
}
if (/alpha done/i.test(bodies[1] || '')) {
  throw new Error('expected second tool card to remain beta slot, got ' + JSON.stringify(bodies));
}
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```