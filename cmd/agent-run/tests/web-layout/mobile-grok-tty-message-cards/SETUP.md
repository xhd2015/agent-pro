# Scenario

**Feature**: grok-tty session message cards match agent-chat UX invariants (roles, bodies, progress separation)

```
seed grok-tty/web_a1e886dbcebb3e2b-shaped events -> web -> session route -> distinct bubbles + compact progress + readable bodies
```

## Preconditions

- `playwright-debug` on PATH (`label: chromium`).
- Seeded `events.jsonl` mirrors the objective grok-tty session shape (two user turns, assistant replies, think/tool progress).
- Frontend renders role labels, non-empty message bodies, and progress cards distinct from chat bubbles.

## Steps

1. Seed `grok-tty/web_a1e886dbcebb3e2b` with representative events.
2. Start `agent-run web` with explicit token.
3. Open session URL and assert message-card UX invariants.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "grok-tty-message-cards"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	sessionID := "web_a1e886dbcebb3e2b"
	workspacePath := filepath.Join(req.TempDir, "grok-message-cards")

	if err := seedGrokTTYMessageCardSession(t, req.Home, sessionID, workspacePath); err != nil {
		return err
	}

	req.Env = append(req.Env,
		"AGENT_RUN_GROK_TTY_GROK_SESSION_ID=",
		"LLM_MOCK_RUN_GROK_COMMAND=",
	)

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/grok-tty/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
await page.locator('[data-testid="chat-active"]').waitFor({ state: 'visible', timeout: 15000 });
const userCards = page.locator('[data-testid="message-item-user"]');
const assistantCards = page.locator('[data-testid="message-item-assistant"]');
await userCards.first().waitFor({ state: 'visible', timeout: 15000 });
await assistantCards.first().waitFor({ state: 'visible', timeout: 15000 });
const userCount = await userCards.count();
const assistantCount = await assistantCards.count();
if (userCount < 2) throw new Error('expected at least 2 user message cards, got ' + userCount);
if (assistantCount < 2) throw new Error('expected at least 2 assistant message cards, got ' + assistantCount);
const userBody = (await userCards.first().locator('.message-body').innerText()).trim();
const assistantBody = (await assistantCards.first().locator('.message-body').innerText()).trim();
if (!userBody.includes('run ls and pwd')) throw new Error('missing first user body: ' + userBody);
if (!assistantBody.includes('pwd')) throw new Error('missing first assistant body: ' + assistantBody);
const userStyle = await userCards.first().evaluate(el => {
  const s = getComputedStyle(el);
  const r = el.getBoundingClientRect();
  return { backgroundColor: s.backgroundColor, alignSelf: s.alignSelf, left: r.left };
});
const assistantStyle = await assistantCards.first().evaluate(el => {
  const s = getComputedStyle(el);
  const r = el.getBoundingClientRect();
  return { backgroundColor: s.backgroundColor, alignSelf: s.alignSelf, left: r.left };
});
const distinctBg = userStyle.backgroundColor !== assistantStyle.backgroundColor;
const distinctAlign = userStyle.left > assistantStyle.left + 20 || assistantStyle.left > userStyle.left + 20;
if (!distinctBg && !distinctAlign) {
  throw new Error('user and assistant cards lack visual distinction: ' + JSON.stringify({ userStyle, assistantStyle }));
}
const roleLabels = await page.locator('.message-role').allTextContents();
if (!roleLabels.some(t => /you/i.test(t)) || !roleLabels.some(t => /agent/i.test(t))) {
  throw new Error('missing You/Agent role labels: ' + JSON.stringify(roleLabels));
}
const progress = page.locator('[data-testid="progress-card"]');
const progressCount = await progress.count();
if (progressCount < 2 || progressCount > 4) {
  throw new Error('expected 2-4 compacted progress cards, got ' + progressCount);
}
const progressStyle = await progress.first().evaluate(el => {
  const s = getComputedStyle(el);
  return { backgroundColor: s.backgroundColor, borderRadius: s.borderRadius };
});
if (userStyle.backgroundColor === progressStyle.backgroundColor) {
  throw new Error('progress card matches user bubble background');
}
const docOverflowX = await page.evaluate(() => Math.max(
  document.documentElement.scrollWidth - document.documentElement.clientWidth,
  document.body.scrollWidth - document.body.clientWidth
));
if (docOverflowX > 2) throw new Error('horizontal overflow ' + docOverflowX + 'px');
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```