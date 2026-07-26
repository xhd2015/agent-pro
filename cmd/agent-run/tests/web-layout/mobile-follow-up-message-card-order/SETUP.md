# Scenario

**Bug fix lock**: live grok-tty follow-up must not jump the first user message card under assistants

```
seed idle turn-1 (user "run ls" + assistant) → /sessions/<id> → composer "what did I say"
→ live timeline keeps chronological users; no assistant before first user
```

## Preconditions

- Web started with grok mock harness (`--agent-runner grok-tty`).
- Open API (`WebTokenMode=omit`).
- `playwright-debug` on PATH.
- Seeded flat idle session with completed turn-1 (user + assistant) so follow-up does not depend on live first-turn bind.

## Steps

1. Seed flat `sessions/follow-up-card-order` (`meta.runner=grok-tty`, `status=idle`) with user `run ls` and assistant `MOCK_REPLY: turn1`.
2. Start grok mock web with open API on a free port (mock prompt = follow-up text).
3. Open `/sessions/follow-up-card-order`; wait `chat-active`; confirm first user visible.
4. Send follow-up `what did I say` via composer (live grok mock run).
5. Poll live timeline (≤60s): chronological users, each prompt once, no non-empty assistant before first user.
6. Reload session URL and re-check the same order invariants.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "follow-up-message-card-order"
	req.WebTokenMode = "omit"
	runner := "grok-tty"
	sessionID := "follow-up-card-order"
	prompt1 := "run ls"
	prompt2 := "what did I say"
	assistantTurn1 := "MOCK_REPLY: turn1"
	marker := layoutGrokAssistantMarker(prompt2)

	if err := seedIdleSessionWithUserAndAssistant(t, req.Home, runner, sessionID, prompt1, assistantTurn1); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt2, marker, 4); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	// Flat SPA route: /sessions/:sessionId (no runner segment).
	sessionPath := "/sessions/" + sessionID
	body := openSeededSessionPage(req.BaseURL, sessionPath) + `
{
  const firstUser = page.locator('[data-testid="message-item-user"]');
  await firstUser.first().waitFor({ state: 'visible', timeout: 15000 });
  const text = (await firstUser.first().locator('.message-body').innerText().catch(() => '')).trim();
  if (!text.includes('run ls') && !'run ls'.includes(text)) {
    throw new Error('seeded first user missing run ls, body=' + JSON.stringify(text));
  }
  const assistantCount = await page.locator('[data-testid="message-item-assistant"]').count();
  if (assistantCount < 1) {
    throw new Error('seeded turn-1 assistant missing');
  }
}
` + sendComposerMessage(prompt2) +
		assertLiveFollowUpMessageCardOrder(prompt1, prompt2, true)

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
