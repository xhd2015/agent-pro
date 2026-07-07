# Scenario

**Bug**: follow-up via composer must not duplicate user timeline bubbles during agent run

```
seed idle grok-tty session + 1 user event -> session page -> composer follow-up -> grok-tty run -> exactly 2 user bubbles
```

## Preconditions

- Web started with grok mock harness (`--agent-runner grok-tty`).
- `playwright-debug` on PATH.
- Seeded session `status=idle` with one prior user message in `events.jsonl`.

## Steps

1. Seed `grok-tty/follow-up-dedupe` with initial user prompt `first layout prompt`.
2. Start grok mock web with open API on a free port.
3. Open session route; send `second follow-up prompt` via composer.
4. While session is running, poll user bubble count every 250ms; fail immediately if count > 2.
5. After run completes, assert exactly two user bubbles; each prompt text appears once.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
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

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertUserMessageCount(1) + sendComposerMessage(followUpPrompt) + assertNoDuplicateUserMessagesDuringRun(2) +
		waitForSessionRunComplete() + assertUserMessageCount(2) + assertDistinctUserPromptsOnce([]string{firstPrompt, followUpPrompt})

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```