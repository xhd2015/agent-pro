# Scenario

**Bug**: terminal button regression — finished chat with known mapping still shows enabled Terminal button

```
finished grok-tty + terminal_session_id + live registry
  -> chat page -> Terminal button visible and enabled
```

## Preconditions

- Same seed as `no-repeat-poll`.
- Terminal affordance must not depend on perpetual `/terminal` polling.

## Steps

1. Seed finished session with known terminal mapping and live registry.
2. Start `agent-run web`.
3. Open session page; wait for Terminal button.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Scenario = "terminal-button-visible"
	seedFinishedKnownTerminalSession(t, req)

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
if (!(await terminalButton.isEnabled())) {
  throw new Error('terminal button disabled for finished chat with live terminal mapping');
}
const statusPillText = await page.locator('.status-pill').first().textContent().catch(() => '');
if (!String(statusPillText || '').toLowerCase().includes('finished')) {
  throw new Error('expected finished status pill, got ' + statusPillText);
}
`)
	return nil
}
```