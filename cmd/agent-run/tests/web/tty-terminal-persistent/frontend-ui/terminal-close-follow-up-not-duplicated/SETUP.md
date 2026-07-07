# Scenario

**Bug**: closing the terminal after the first tty response leaves a stale SSE
stream that replays the next user message.

```
running grok-tty chat
  -> open Terminal while the first turn is running
  -> first assistant response arrives while terminal modal is open
  -> close Terminal
  -> send follow-up from chat composer
  -> follow-up user message appears exactly once
```

## Preconditions

- The session is created through the real web API with grok mock harness.
- The mock hook sleeps before completing the first response while the terminal
  modal is open.

## Steps

1. Create a running `grok-tty` web session.
2. Wait for the tty registry entry so the Terminal button can attach.
3. Open the generated chat route in Playwright.
4. Open the terminal modal and wait for the delayed response text.
5. Close the terminal modal.
6. Send `what did I say?` from the chat composer.
7. Assert the visible follow-up user message count remains exactly one.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	createRunningWebGrokTTYSessionThroughAPI(t, req)
	waitForAnyRegistryID(t, req, 3_000_000_000)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
await terminalButton.click();
await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 15000 });
await page.waitForFunction(() => {
  const text = document.body ? document.body.textContent || '' : '';
  return text.includes('delayed terminal run completed');
}, null, { timeout: 10000 });
await page.getByRole('button', { name: /close terminal/i }).click();
await page.getByTestId('chat-active').waitFor({ state: 'visible', timeout: 15000 });

const followUp = 'what did I say?';
const userBodies = async () => await page.locator('[data-testid="message-item-user"] .message-body').allTextContents();
console.log('before follow-up bodies', JSON.stringify(await userBodies()));
await page.getByTestId('composer-input').fill(followUp);
await page.getByTestId('send-button').click();
for (const delay of [500, 1500, 3000]) {
  await page.waitForTimeout(delay);
  const bodies = await userBodies();
  const count = bodies.filter((text) => text === followUp).length;
  console.log('follow-up bodies after delay', delay, JSON.stringify(bodies), 'count', count);
  if (count !== 1) {
    throw new Error('follow-up should appear exactly once after closing terminal; count=' + count + ' bodies=' + JSON.stringify(bodies));
  }
}
`)
	return nil
}
```