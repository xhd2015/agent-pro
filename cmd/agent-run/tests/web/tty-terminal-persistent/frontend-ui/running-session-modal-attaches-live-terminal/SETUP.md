# Scenario

**Bug**: terminal modal says unavailable during a running tty turn and only shows exited content after finish

```
web-created grok-tty running session + live server-side PTY
  -> click Terminal before assistant response finishes
  -> modal renders live GROK_TTY_BANNER / prompt
  -> modal does not show terminal unavailable or exited
```

## Preconditions

- The session is created through the real web API with grok mock harness.
- The mock hook prints `GROK_TTY_BANNER` and sleeps before the final assistant response.
- The test opens the terminal modal while the turn is still running.

## Steps

1. Create a slow running `grok-tty` web session.
2. Wait for the tty registry entry so the backend PTY exists.
3. Open the generated chat route in Playwright.
4. Click the terminal button while the turn is running.
5. Assert live terminal content is visible and unavailable/exited text is absent.

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
const surface = page.getByTestId('terminal-surface');
await surface.waitFor({ state: 'visible', timeout: 15000 });
const terminalText = async () => await page.locator('body').textContent();
try {
  await page.waitForFunction(() => {
    const text = document.body ? document.body.textContent || '' : '';
    return text.includes('GROK_TTY_BANNER') || text.includes('Grok');
  }, null, { timeout: 2000 });
} catch (err) {
  const text = await terminalText();
  throw new Error('terminal modal did not show live running tty content; visible text=' + JSON.stringify(text));
}
const text = await terminalText();
console.log('running terminal modal text', JSON.stringify(text));
if (String(text || '').includes('terminal unavailable')) {
  throw new Error('terminal modal reported unavailable while tty turn was running: ' + text);
}
if (String(text || '').toLowerCase().includes('exited')) {
  throw new Error('terminal modal showed exited terminal during active tty turn: ' + text);
}
if (String(text || '').includes('delayed terminal run completed')) {
  throw new Error('assistant response finished before live terminal assertion; test no longer covers running modal attach');
}
`)
	return nil
}
```