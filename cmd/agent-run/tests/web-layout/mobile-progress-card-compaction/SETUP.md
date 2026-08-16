# Scenario

**Feature**: progress cards compact duplicate tool/think events and stay visually distinct from chat bubbles

```
seed user + duplicate tool_call + think + assistant -> session page -> fewer progress cards than raw events
```

## Preconditions

- `playwright-debug` on PATH (`label: chromium`).
- Seeded `events.jsonl` includes duplicate `tool_call_id` rows and consecutive `think` events.
- Frontend compacts progress timeline and renders `progress-card` with role labels.

## Steps

1. Seed `fake-opencode/layout-progress-compact` with duplicate tool calls and think events.
2. Start `agent-run web` with explicit token.
3. Open session URL; assert compacted progress cards and distinct styling from message bubbles.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "progress-card-compaction"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-progress-compact"
	workspacePath := filepath.Join(req.TempDir, "progress-compact")

	if err := seedProgressCompactionSession(t, req.Home, runner, sessionID, workspacePath); err != nil {
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
if (count < 2) throw new Error('expected at least 2 compacted progress cards, got ' + count);
if (count > 4) throw new Error('expected at most 4 compacted progress cards, got ' + count);
const labels = await page.locator('.progress-card-label').allTextContents();
if (!labels.some(t => /thinking/i.test(t))) throw new Error('missing Thinking label, got ' + JSON.stringify(labels));
if (!labels.some(t => /tool/i.test(t))) throw new Error('missing Tool label, got ' + JSON.stringify(labels));
const userBubble = page.locator('[data-testid="message-item-user"]').first();
const progressCard = progress.first();
const userStyle = await userBubble.evaluate(el => {
  const s = getComputedStyle(el);
  return { backgroundColor: s.backgroundColor, borderRadius: s.borderRadius };
});
const progressStyle = await progressCard.evaluate(el => {
  const s = getComputedStyle(el);
  return { backgroundColor: s.backgroundColor, borderRadius: s.borderRadius };
});
if (userStyle.backgroundColor === progressStyle.backgroundColor) {
  throw new Error('progress card matches user bubble background: ' + JSON.stringify({ userStyle, progressStyle }));
}
const bodyMaxHeight = await progressCard.locator('.progress-card-body').evaluate(el => getComputedStyle(el).maxHeight);
if (!bodyMaxHeight || bodyMaxHeight === 'none') {
  throw new Error('progress-card-body missing max-height clamp');
}
const progressBoxes = await progress.all();
if (progressBoxes.length < 2) throw new Error('expected at least 2 progress cards for ordering check');
const lastProgressLabel = await progressBoxes[progressBoxes.length - 1].locator('.progress-card-label').innerText();
if (!/tool/i.test(lastProgressLabel)) {
  throw new Error('expected merged tool card after intervening think, got label: ' + lastProgressLabel);
}
const thinkIdx = labels.findIndex(t => /thinking/i.test(t));
const toolIdx = labels.findIndex(t => /tool/i.test(t));
if (thinkIdx < 0 || toolIdx < 0 || thinkIdx > toolIdx) {
  throw new Error('expected Thinking card before Tool card, labels=' + JSON.stringify(labels));
}
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```