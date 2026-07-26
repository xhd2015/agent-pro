# Scenario

**Bug**: terminal button hidden until assistant response completes on running tty session

```
web-created grok-tty running session + terminal_session_id session-N
  -> open /sessions/grok-tty/web_*
  -> terminal button visible before assistant response finishes
```

## Preconditions

- Session created through real web API with slow grok mock hook.
- TTY registry entry exists before assistant response completes.

## Steps

1. Create a new `grok-tty` web session via API.
2. Wait for tty registry entry.
3. Open generated chat route in Playwright.
4. Assert terminal button visible before delayed response text appears.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "ui"
	createRunningWebGrokTTYSessionThroughAPI(t, req)
	waitForAnyRegistryID(t, req, 3_000_000_000)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
const bodyText = await page.locator('body').textContent() || '';
if (bodyText.includes('delayed terminal run completed')) {
  throw new Error('assistant response finished before terminal icon assertion');
}
`)
	return nil
}
```