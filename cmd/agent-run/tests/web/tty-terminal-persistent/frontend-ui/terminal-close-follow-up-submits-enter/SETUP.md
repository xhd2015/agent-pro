# Scenario

**Bug**: after closing a terminal that displayed the first response, a chat
follow-up does not produce a visible assistant response.

```
new grok-tty chat
  -> user opens Terminal while first turn is running
  -> first response appears in Terminal and chat
  -> user closes Terminal
  -> user sends follow-up from chat composer
  -> backend TTY receives follow-up plus Enter
  -> chat page shows the follow-up response
```

## Preconditions

- The session is created through the real web API with grok mock two-turn hook.
- The mock hook stays alive after the first answer and emits `FOLLOWUP_RESPONSE`
  after reading a second PTY line.

## Steps

1. Start a `grok-tty` web session with `one word of France capital`.
2. Open the terminal modal while the first turn is still running.
3. Wait for the first terminal response `Paris`.
4. Close the terminal modal.
5. Send `What did I say` from the chat composer.
6. Reopen the terminal.
7. Assert `FOLLOWUP_RESPONSE` appears in chat.

```go
import (
	"net/http"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "ui"
	prompt := "one word of France capital"
	setGrokMockHook(t, req, llmMockGrokTTYTwoTurnHook(prompt, persistentGrokMockUUID))
	restartWebWithGrokMock(t, req)
	body := `{"runner":"grok-tty","prompt":"one word of France capital"}`
	status, respBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions", req.WebToken, "application/json", body)
	if status != http.StatusAccepted {
		t.Fatalf("create two-turn web grok-tty session status=%d body=%s", status, respBody)
	}
	created := decodeJSONBody(t, respBody)
	session, _ := created["session"].(map[string]any)
	sessionID, _ := session["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("create response missing session_id: %s", respBody)
	}
	req.Runner = "grok-tty"
	req.ChatSessionID = sessionID
	waitForAnyRegistryID(t, req, 3_000_000_000)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
await terminalButton.click();
await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 15000 });
await page.waitForFunction(() => {
  const text = document.body ? document.body.textContent || '' : '';
  return text.includes('Response: Paris');
}, null, { timeout: 10000 });
await page.getByRole('button', { name: /close terminal/i }).click();
await page.getByTestId('chat-active').waitFor({ state: 'visible', timeout: 15000 });

await page.getByTestId('composer-input').fill('What did I say');
await page.getByTestId('send-button').click();
await page.waitForTimeout(5000);

const chatText = await page.locator('body').textContent() || '';
if (!chatText.includes('FOLLOWUP_RESPONSE')) {
  await page.getByRole('button', { name: /terminal/i }).click();
  await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 15000 });
  let terminalShowsResponse = false;
  try {
    await page.waitForFunction(() => {
      const text = document.body ? document.body.textContent || '' : '';
      return text.includes('FOLLOWUP_RESPONSE');
    }, null, { timeout: 3000 });
    terminalShowsResponse = true;
  } catch (err) {
    terminalShowsResponse = false;
  }
  const terminalText = await page.locator('body').textContent() || '';
  throw new Error(
    'chat did not show follow-up response after closing terminal; ' +
    'terminalShowsResponse=' + terminalShowsResponse + ' visibleText=' + JSON.stringify(terminalText)
  );
}

await page.getByRole('button', { name: /terminal/i }).click();
await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 15000 });
try {
  await page.waitForFunction(() => {
    const text = document.body ? document.body.textContent || '' : '';
    return text.includes('FOLLOWUP_RESPONSE');
  }, null, { timeout: 6000 });
} catch (err) {
  const text = document.body ? document.body.textContent || '' : '';
  throw new Error('follow-up was not submitted to the backend TTY; terminal text=' + JSON.stringify(text));
}
`)
	return nil
}
```