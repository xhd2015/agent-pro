# Scenario

**Bug**: with the real `codex` TUI, closing a terminal after the first response
can leave typed terminal input behind. The next chat follow-up is then appended
to that stale terminal input or the chat stores raw terminal scrollback as the
assistant response.

```
real codex-tty web chat
  -> first prompt returns Paris
  -> user opens Terminal and types a stale terminal draft without Enter
  -> user closes Terminal
  -> user sends chat follow-up "what did I say?"
  -> backend clears stale terminal draft, submits only the follow-up, and keeps
     raw TUI scrollback out of chat messages
```

## Preconditions

- Real `codex` CLI must be on `PATH`; otherwise this leaf skips.
- The test does not install or shadow a fake `codex` binary; it points the
  TTY command override at the real `codex` binary with Codex CLI flags.
- Playwright drives the real chat page and terminal modal.

## Steps

1. Start a real `codex-tty` web session with `one word of France capital`.
2. Open the terminal modal while the first turn is running.
3. Wait for the real Codex output to contain `Paris`.
4. Type `Explain this codebase` into the terminal without pressing Enter.
5. Close the terminal.
6. Send `what did I say?` from the chat composer.
7. Reopen the terminal and assert the stale terminal draft was not submitted
   together with the follow-up.
8. Assert the chat does not show raw terminal control/session output as the
   assistant response.

```go
import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	codexPath, err := exec.LookPath("codex")
	if err != nil {
		t.Skipf("codex not found in PATH: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(req.AgentRun), "codex")); err == nil {
		t.Fatalf("real codex leaf must not use fake codex fixture binary")
	}
	req.Env = withoutEnvKey(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND")
	req.Env = append(req.Env, "AGENT_RUN_CODEX_TTY_COMMAND="+
		strconv.Quote(codexPath)+" --dangerously-bypass-approvals-and-sandbox --cd "+strconv.Quote(req.RepoRoot))
	restartWebWithCurrentEnv(t, req)

	req.Mode = "ui"
	req.Runner = "codex-tty"

	body := `{"runner":"codex-tty","prompt":"one word of France capital"}`
	status, respBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions", req.WebToken, "application/json", body)
	if status != http.StatusAccepted {
		t.Fatalf("create real codex-tty session status=%d body=%s", status, respBody)
	}
	created := decodeJSONBody(t, respBody)
	session, _ := created["session"].(map[string]any)
	sessionID, _ := session["session_id"].(string)
	if sessionID == "" {
		t.Fatalf("create response missing session_id: %s", respBody)
	}
	req.ChatSessionID = sessionID
	waitForAnyRegistryID(t, req, 20_000_000_000)

	req.PlaywrightScript = sessionBrowserScript(req, `
const visibleSnippet = (text) => String(text || '')
  .replace(/\\.xterm[^{}]*\\{[^{}]*\\}/g, ' ')
  .replace(/\\s+/g, ' ')
  .slice(-3000);

const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 30000 });
await terminalButton.click();
await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 30000 });

try {
  await page.waitForFunction(() => {
    const bodyText = document.body ? document.body.textContent || '' : '';
    const assistantText = Array.from(document.querySelectorAll('[data-testid="message-item-assistant"]'))
      .map((node) => node.textContent || '')
      .join('\\n');
    const text = bodyText + '\\n' + assistantText;
    return text.includes('Paris') || text.includes('[Terminal exited]') || text.includes('To continue this session');
  }, null, { timeout: 180000 });
} catch (err) {
  const bodyText = await page.locator('body').textContent() || '';
  const terminalText = await page.locator('[data-testid="terminal-surface"]').textContent().catch(() => '') || '';
  const assistantText = await page.locator('[data-testid="message-item-assistant"]').allTextContents().catch(() => []);
  throw new Error('real Codex did not produce visible Paris before stale-input step; body=' + JSON.stringify(visibleSnippet(bodyText)) + ' terminal=' + JSON.stringify(visibleSnippet(terminalText)) + ' assistant=' + JSON.stringify(assistantText.map(visibleSnippet)));
}
const firstTurnText = await page.locator('body').textContent() || '';
if (!firstTurnText.includes('Paris')) {
  throw new Error('real Codex terminal exited or printed resume footer before answering first prompt; visibleText=' + JSON.stringify(visibleSnippet(firstTurnText)));
}

const helper = page.locator('[data-testid="terminal-surface"] .xterm-helper-textarea');
await helper.waitFor({ state: 'attached', timeout: 30000 });
await helper.focus();
await page.keyboard.type('Explain this codebase');
await page.waitForTimeout(500);

await page.getByRole('button', { name: /close terminal/i }).click();
await page.getByTestId('chat-active').waitFor({ state: 'visible', timeout: 30000 });

await page.getByTestId('composer-input').fill('what did I say?');
await page.getByTestId('send-button').click();
await page.waitForTimeout(12000);

const assistantText = await page.locator('[data-testid="message-item-assistant"]').allTextContents();
const assistantJoined = assistantText.join('\\n');
const rawTerminalMarkers = [
  '"type":"session_id"',
  'OpenAI Codex',
  'model:',
  'directory:',
  'Explain this codebase',
  'whatdidIsay',
  '[Terminal exited]',
  'esc to interrupt'
];
const foundRawMarker = rawTerminalMarkers.find((marker) => assistantJoined.includes(marker));
if (foundRawMarker) {
  throw new Error('chat assistant message contains raw real Codex terminal output marker ' + JSON.stringify(foundRawMarker) + ': ' + JSON.stringify(assistantJoined));
}

await terminalButton.click();
await page.getByTestId('terminal-surface').waitFor({ state: 'visible', timeout: 30000 });
await page.waitForTimeout(2000);
const terminalText = await page.locator('[data-testid="terminal-surface"]').textContent() || '';
const compact = terminalText.replace(/\\s+/g, '');
if (compact.includes('ExplainthiscodebasewhatdidIsay') || compact.includes('Explainthiscodebase›whatdidIsay')) {
  throw new Error('real Codex terminal submitted stale draft together with chat follow-up: ' + JSON.stringify(terminalText));
}
if (terminalText.includes('Explain this codebase') && terminalText.includes('what did I say?')) {
  throw new Error('real Codex terminal still contains stale draft when chat follow-up was sent: ' + JSON.stringify(terminalText));
}
`)
	return nil
}

func withoutEnvKey(env []string, key string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func restartWebWithCurrentEnv(t *testing.T, req *Request) {
	t.Helper()
	if req.WebCmd != nil && req.WebCmd.Process != nil {
		_ = req.WebCmd.Process.Kill()
		_ = req.WebCmd.Wait()
	}
	startAgentRunWeb(t, req)
}
```
