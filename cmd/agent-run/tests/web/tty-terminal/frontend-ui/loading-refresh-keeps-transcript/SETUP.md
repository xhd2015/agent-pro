# Scenario

**Feature**: refreshing terminal/session state does not hide loaded transcript

```
chat transcript visible -> trigger terminal status/session refresh -> transcript stays visible while refresh is pending
```

## Preconditions

- Existing chat contains loaded messages.
- Loading indicators must be additive and not replace main content.

## Steps

1. Seed tty session and live registry.
2. Open chat page and observe transcript.
3. Trigger a refresh by opening/closing terminal affordance and waiting through polling.
4. Verify transcript text remains visible.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "ui-refresh-keeps-chat"
	req.RegistryTranscript = "refresh-terminal\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
const transcript = page.getByText('assistant keeps transcript');
await transcript.waitFor({ state: 'visible', timeout: 15000 });
await page.getByRole('button', { name: /terminal/i }).click();
await page.getByRole('dialog').waitFor({ state: 'visible', timeout: 15000 });
await page.getByRole('button', { name: /close|dismiss/i }).click();
await page.waitForTimeout(1200);
if (!(await transcript.isVisible())) {
  throw new Error('transcript hidden during terminal/session refresh');
}
`)
	return nil
}
```
