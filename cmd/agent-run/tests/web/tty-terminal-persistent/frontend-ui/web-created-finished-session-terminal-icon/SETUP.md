# Scenario

**Bug**: generated grok-tty web session route lacks terminal icon after first turn

```
web-created grok-tty session -> finished + terminal_session_id session-N
browser opens generated /sessions/grok-tty/web_* route -> terminal button visible
```

## Preconditions

- Web started with grok mock harness.
- Session created through real web API (not seeded fixture).

## Steps

1. Create a new `grok-tty` session via web API.
2. Wait for session finish and terminal mapping.
3. Open generated route in Playwright.
4. Assert terminal button visible and enabled.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	createWebGrokTTYSessionThroughAPI(t, req)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
const terminalStatus = page.getByTestId('terminal-status');
const statusText = await terminalStatus.textContent().catch(() => '');
if (statusText && statusText.toLowerCase().includes('unavailable')) {
  throw new Error('terminal status unavailable for generated finished grok-tty session: ' + terminalStatus.text);
}
if (!(await terminalButton.isEnabled())) {
  throw new Error('terminal button disabled for generated finished grok-tty session');
}
`)
	return nil
}
```