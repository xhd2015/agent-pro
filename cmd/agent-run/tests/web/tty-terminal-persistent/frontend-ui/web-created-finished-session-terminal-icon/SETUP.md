# Scenario

**Bug**: generated codex-tty web session route lacks terminal icon after first turn

```
web-created codex-tty session -> finished + terminal_session_id session-N
browser opens generated /sessions/codex-tty/web_* route -> terminal button visible
```

## Preconditions

- This uses the real `agent-run web` API to create the session.
- A fake `codex` binary is placed on PATH so the run is deterministic.
- The final assertion is made in Playwright against the generated chat route.

## Steps

1. Create a new codex-tty session via web API.
2. Wait for finished session metadata with `terminal_session_id`.
3. Open the generated chat route in Playwright.
4. From browser context, fetch `/terminal` and assert it is available.
5. Assert a visible enabled terminal button exists in the top bar.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	createWebCodexTTYSessionThroughAPI(t, req)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalStatus = await page.evaluate(async ({ runner, sessionId, token }) => {
  const res = await fetch('/api/agent-run/sessions/' + encodeURIComponent(runner) + '/' + encodeURIComponent(sessionId) + '/terminal', {
    headers: { Authorization: 'Bearer ' + token },
  });
  const text = await res.text();
  let json = null;
  try { json = JSON.parse(text); } catch {}
  return { status: res.status, text, json };
}, { runner: `+jsQuote(req.Runner)+`, sessionId: `+jsQuote(req.ChatSessionID)+`, token: `+jsQuote(req.WebToken)+` });
console.log('terminal status payload', JSON.stringify(terminalStatus));
if (terminalStatus.status !== 200) {
  throw new Error('terminal status HTTP ' + terminalStatus.status + ': ' + terminalStatus.text);
}
if (!terminalStatus.json || terminalStatus.json.terminal_session_id !== `+jsQuote(req.TerminalSessionID)+`) {
  throw new Error('terminal status did not echo generated terminal_session_id `+req.TerminalSessionID+`: ' + terminalStatus.text);
}
if (!terminalStatus.json.available) {
  throw new Error('terminal status unavailable for generated finished codex-tty session: ' + terminalStatus.text);
}
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
if (!(await terminalButton.isEnabled())) {
  throw new Error('terminal button disabled for generated finished codex-tty session');
}
`)
	return nil
}
```
