# Scenario

**Bug**: terminal modal shows raw stream characters instead of a real terminal

```
ptywrap sends control JSON + ANSI terminal bytes
  -> browser modal
  -> xterm-like terminal surface renders READY without raw escape text
  -> typed input is sent back to the PTY
```

## Preconditions

- Session metadata stores `terminal_session_id:"session-1"`.
- The fake ptywrap server sends a `session_id` control frame before terminal
  bytes, matching real ptywrap behavior.
- Terminal bytes include ANSI SGR color controls.

## Steps

1. Write finished mapped session metadata.
2. Write registry entry pointing at a ptywrap-like websocket server.
3. Open chat page in Playwright and click the terminal button.
4. Assert the modal contains an actual terminal emulator element, interprets ANSI
   bytes, hides control JSON, and sends typed input through the websocket.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RegistryTranscript = "\x1b[31mREADY\x1b[0m\r\n"
	listenAddr := startControlFramePtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	req.Mode = "ui"
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
await terminalButton.click();
const surface = page.getByTestId('terminal-surface');
await surface.waitFor({ state: 'visible', timeout: 15000 });
await page.locator('[data-testid="terminal-surface"] .xterm').waitFor({ state: 'visible', timeout: 15000 });
const rows = page.locator('[data-testid="terminal-surface"] .xterm-rows');
await page.waitForFunction(() => {
  const el = document.querySelector('[data-testid="terminal-surface"] .xterm-rows');
  return el && el.textContent && el.textContent.includes('READY');
}, null, { timeout: 15000 });
const initialText = await rows.textContent();
console.log('terminal emulator initial text', JSON.stringify(initialText));
if (String(initialText || '').includes('\\u001b') || String(initialText || '').includes('[31m') || String(initialText || '').includes('[0m')) {
  throw new Error('modal rendered raw ANSI escape characters instead of terminal output: ' + initialText);
}
if (String(initialText || '').includes('"type":"session_id"') || String(initialText || '').includes('session_id')) {
  throw new Error('modal rendered ptywrap control JSON instead of terminal output: ' + initialText);
}
await page.locator('[data-testid="terminal-surface"] .xterm-helper-textarea').focus();
await page.keyboard.type('abc');
await page.keyboard.press('Enter');
await page.waitForFunction(() => {
  const el = document.querySelector('[data-testid="terminal-surface"] .xterm-rows');
  const text = el && el.textContent ? el.textContent : '';
  return text.includes('echo:abc') || text.includes('echo:abc\\r');
}, null, { timeout: 15000 });
const interactiveText = await rows.textContent();
console.log('terminal emulator interactive text', JSON.stringify(interactiveText));
`)
	return nil
}
```
