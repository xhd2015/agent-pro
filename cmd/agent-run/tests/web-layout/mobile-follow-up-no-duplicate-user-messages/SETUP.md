# Scenario

**Bug**: follow-up via composer must not duplicate user timeline bubbles during agent run

```
seed idle grok-tty session + 1 user event -> flat session page -> composer follow-up -> grok-tty run -> exactly 2 user bubbles
```

## Preconditions

- Web started with grok mock harness (`--agent-runner grok-tty`).
- `playwright-debug` on PATH.
- Seeded session `status=idle` with one prior user message in flat `sessions/<id>/events.jsonl`.

## Steps

1. Seed flat `sessions/follow-up-dedupe` (meta.runner=`grok-tty`) with initial user prompt `first layout prompt`.
2. Start grok mock web with open API on a free port.
3. Open flat session route `/sessions/follow-up-dedupe`; send `second follow-up prompt` via composer.
4. While session is running, poll user bubble count every 250ms; fail immediately if count > 2.
5. Poll until user bubble count is exactly 2 (no idle/assistant requirement — session may stay running on mock bind races).
6. Assert exactly two user bubbles; each prompt text appears once.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "follow-up-dedupe"
	req.WebTokenMode = "omit"
	runner := "grok-tty"
	sessionID := "follow-up-dedupe"
	firstPrompt := "first layout prompt"
	followUpPrompt := "second follow-up prompt"
	marker := layoutGrokAssistantMarker(followUpPrompt)

	if err := seedIdleSessionWithUserMessage(t, req.Home, runner, sessionID, firstPrompt); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, followUpPrompt, marker, 6); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	// Flat SPA route: /sessions/:sessionId (no runner segment).
	sessionPath := "/sessions/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertUserMessageCount(1) + sendComposerMessage(followUpPrompt) + assertNoDuplicateUserMessagesDuringRun(2) +
		waitForUserMessageCount(2) + assertUserMessageCount(2) + assertDistinctUserPromptsOnce([]string{firstPrompt, followUpPrompt})

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```
