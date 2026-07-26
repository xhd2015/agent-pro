# Scenario

**Feature**: chat UI shows assistant message without phased coalescing rows

```
POST grok-tty -> open chat page -> one assistant bubble for stored message
```

## Steps

1. Create finished `grok-tty` session via grok mock web harness.
2. Run Playwright script counting assistant message bubbles.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prompt := "ui assistant parity"
	marker := "WEB_CLI_ASSISTANT_MARKER"
	startGrokTTYWebMockEnv(t, req, prompt, marker, 2)
	startAgentRunWeb(t, req)
	req.Runner = "grok-tty"
	req.Prompt = prompt
	sessionID, _, _ := postCreateSession(t, req, req.Runner, req.Prompt)
	req.SessionID = sessionID
	waitForSessionStatus(t, req, req.Runner, sessionID, "finished", 45*time.Second)
	req.PlaywrightScript = sessionBrowserScript(req, `
const bubbles = await page.locator('[data-testid="assistant-message"]').count();
if (bubbles !== 1) throw new Error('expected 1 assistant bubble, got ' + bubbles);
const phased = await page.locator('[data-phase]').count();
if (phased !== 0) throw new Error('phased DOM rows present: ' + phased);
`)
	return nil
}
```