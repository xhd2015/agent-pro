# Scenario

**Bug**: terminal button appears only after the tty agent response finishes

```
web-created codex-tty running session + terminal_session_id session-N
  -> open /sessions/codex-tty/web_*
  -> terminal button visible before assistant completion
```

## Preconditions

- This uses the real `agent-run web` session creation API.
- A fake `codex` binary keeps the run active for several seconds after the
  terminal is created.
- The chat route is opened while the run is still active.

## Steps

1. Create a new `codex-tty` web session via API.
2. Wait only for `terminal_session_id`, not for the assistant response.
3. Open the generated chat route in Playwright.
4. Assert the terminal button is visible quickly while the delayed assistant
   response is still absent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	createRunningWebCodexTTYSessionThroughAPI(t, req)
	req.PlaywrightScript = sessionBrowserScript(req, `
const started = Date.now();
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 300 });
const elapsed = Date.now() - started;
const bodyText = await page.locator('body').textContent();
console.log('running terminal button elapsed', elapsed);
console.log('running page text before response', JSON.stringify(bodyText));
if (String(bodyText || '').includes('delayed terminal run completed')) {
  throw new Error('assistant response finished before terminal icon assertion; test no longer covers running-state icon timing');
}
if (!(await terminalButton.isEnabled())) {
  throw new Error('terminal button was visible but disabled while tty session was running');
}
`)
	return nil
}
```
