# Scenario

**Bug**: terminal modal renders ptywrap `session_id` control JSON instead of terminal output

```
mapped terminal session-1 -> websocket opens modal
ptywrap sends {"type":"session_id","session_id":"session-1"} control frame
modal displays terminal bytes only, not control JSON
```

## Preconditions

- Session metadata stores `terminal_session_id:"session-1"`.
- The fake ptywrap server only sends terminal output if the websocket attaches
  with `?session_id=session-1`.
- The fake ptywrap server sends the same initial JSON control frame as real
  ptywrap before binary terminal bytes.

## Steps

1. Write finished mapped session metadata.
2. Write registry entry pointing at ptywrap-like test websocket server.
3. Open chat page in Playwright.
4. Click terminal button.
5. Assert modal shows terminal transcript and does not show JSON control frame.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RegistryTranscript = "mapped-terminal-ready\n"
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
await page.waitForTimeout(500);
const text = await surface.textContent();
console.log('terminal surface text', JSON.stringify(text));
if (!String(text || '').includes('mapped-terminal-ready')) {
  throw new Error('terminal transcript missing from modal: ' + text);
}
if (String(text || '').includes('"type":"session_id"') || String(text || '').includes('session_id')) {
  throw new Error('modal rendered ptywrap session_id control JSON: ' + text);
}
if (String(text || '').includes('created-unmapped')) {
  throw new Error('terminal websocket attached without mapped session_id: ' + text);
}
`)
	return nil
}
```
