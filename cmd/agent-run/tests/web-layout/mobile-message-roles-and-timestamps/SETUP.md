# Scenario

**Feature**: session chat UI shows role-distinct bubbles and per-message timestamps

```
seed events(user+assistant w/ timestamp) -> web -> session route -> role testids + message-timestamp + distinct styles
```

## Preconditions

- `playwright-debug` on PATH (`label: chromium`).
- Seeded `events.jsonl` includes `timestamp` on both message events.
- Frontend renders `message-item-user` / `message-item-assistant` with different alignment or color.

## Steps

1. Seed `fake-opencode/layout-roles-ts` with user and assistant messages (fixed timestamps).
2. Start `agent-run web` with explicit token.
3. Open session URL and compare bubble computed styles and timestamp visibility.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "roles-timestamps"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-roles-ts"
	workspacePath := filepath.Join(req.TempDir, "roles-demo")
	userTS := int64(1719691200000)
	assistantTS := int64(1719691234567)

	if err := seedRoleTimelineSession(t, req.Home, runner, sessionID, workspacePath, userTS, assistantTS); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const userBubble = page.locator('[data-testid="message-item-user"]');
const assistantBubble = page.locator('[data-testid="message-item-assistant"]');
await userBubble.waitFor({ state: 'visible', timeout: 15000 });
await assistantBubble.waitFor({ state: 'visible', timeout: 15000 });
const userStyle = await userBubble.evaluate(el => {
  const s = getComputedStyle(el);
  return { textAlign: s.textAlign, backgroundColor: s.backgroundColor, alignSelf: s.alignSelf };
});
const assistantStyle = await assistantBubble.evaluate(el => {
  const s = getComputedStyle(el);
  return { textAlign: s.textAlign, backgroundColor: s.backgroundColor, alignSelf: s.alignSelf };
});
const sameAlign = userStyle.textAlign === assistantStyle.textAlign && userStyle.alignSelf === assistantStyle.alignSelf;
const sameBg = userStyle.backgroundColor === assistantStyle.backgroundColor;
if (sameAlign && sameBg) {
  throw new Error('user and assistant bubbles look identical: ' + JSON.stringify({ userStyle, assistantStyle }));
}
const timestamps = page.locator('[data-testid="message-timestamp"]');
const tsCount = await timestamps.count();
if (tsCount < 2) {
  throw new Error('expected at least 2 message-timestamp elements, got ' + tsCount);
}
for (let i = 0; i < tsCount; i++) {
  const text = (await timestamps.nth(i).innerText()).trim();
  if (!text) throw new Error('empty message-timestamp at index ' + i);
}
const userText = await userBubble.innerText();
if (!userText.includes('You said hi')) {
  throw new Error('user message text mismatch: ' + userText);
}
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```