# Scenario

**Feature**: chat UI shows error card when grok session bind fails

```
failure bind -> events.jsonl error "Cannot resolve session id: ..."
  -> chat page renders data-testid="error-card"
```

## Steps

1. Configure failure binding env and start web.
2. POST `grok-tty` session and wait for `finished`.
3. Run Playwright script asserting error card with resolve error prefix.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "error-card-on-bind-failure"
	configureBindingFailureEnv(t, req, "ui error card probe")
	startWebGrokSession(t, req)
	waitForSessionComplete(t, req, req.Runner, req.SessionID, bindingFailureFinishTimeout)
	req.PlaywrightScript = sessionBrowserScript(req, `
const card = page.locator('[data-testid="error-card"]').first();
await card.waitFor({ state: 'visible', timeout: 15000 });
const text = await card.innerText();
if (!text.includes('Cannot resolve session id')) {
  throw new Error('error card missing resolve error prefix, got: ' + text);
}
const assistant = await page.locator('[data-testid="assistant-message"]').count();
if (assistant !== 0) {
  throw new Error('expected no assistant bubble fallback, got ' + assistant);
}
`)
	return nil
}
```