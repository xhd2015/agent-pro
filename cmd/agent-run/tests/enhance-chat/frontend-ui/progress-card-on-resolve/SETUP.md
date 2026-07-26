# Scenario

**Feature**: chat UI shows progress card for resolve think event

```
successful bind -> events.jsonl think "Resolve session id..."
  -> chat page renders data-testid="progress-card"
```

## Steps

1. Configure success binding env and start web.
2. POST `grok-tty` session and wait for `finished`.
3. Run Playwright script asserting progress card visible with resolve text.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "progress-card-on-resolve"
	configureBindingSuccessEnv(t, req, "ui progress card probe", enhanceChatSuccessMarker)
	startWebGrokSession(t, req)
	waitForSessionStatus(t, req, req.Runner, req.SessionID, "finished", 45*time.Second)
	req.PlaywrightScript = sessionBrowserScript(req, `
const cards = page.locator('[data-testid="progress-card"]');
await cards.first().waitFor({ state: 'visible', timeout: 15000 });
const count = await cards.count();
if (count < 1) throw new Error('expected at least 1 progress card, got ' + count);
const text = await cards.first().innerText();
if (!text.includes('Resolve session id')) {
  throw new Error('progress card missing resolve text, got: ' + text);
}
`)
	return nil
}
```