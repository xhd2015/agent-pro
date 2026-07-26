# Scenario

**Bug**: finished web tty chat should still show terminal icon when terminal is available

```
finished codex-tty chat -> GET /terminal available true
chat page -> visible enabled Terminal button
```

## Preconditions

- Terminal status is expected to resolve via `terminal_session_id`.

## Steps

1. Start fake ptywrap server.
2. Write finished mapped session metadata.
3. Write live registry for `session-1`.
4. Open the chat page and wait for the terminal button.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.RegistryTranscript = "ui-finished-terminal-ready\n"
	listenAddr := startMappedPtywrap(t, req)
	writeMappedSessionFixture(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
if (!(await terminalButton.isEnabled())) throw new Error('terminal button disabled for finished chat with live terminal');
const statusPillText = await page.locator('.status-pill').first().textContent().catch(() => '');
if (!String(statusPillText || '').toLowerCase().includes('finished')) {
  throw new Error('expected finished status pill, got ' + statusPillText);
}
`)
	return nil
}
```
