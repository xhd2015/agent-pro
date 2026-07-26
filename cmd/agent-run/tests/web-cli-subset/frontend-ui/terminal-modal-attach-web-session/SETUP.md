# Scenario

**Feature**: terminal modal attaches to web-created grok-tty session PTY

```
POST grok-tty -> finished -> open chat -> click terminal -> xterm receives GROK_TTY_BANNER bytes
```

## Steps

1. Create web `grok-tty` session with grok mock harness.
2. Playwright opens terminal modal and waits for PTY banner text.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	createWebGrokTTYSession(t, req, "modal attach probe")
	waitForSessionStatus(t, req, req.Runner, req.ChatSessionID, "finished", 60*time.Second)
	req.SessionID = req.ChatSessionID
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByRole('button', { name: /terminal/i }).click();
await page.waitForSelector('.xterm-rows', { timeout: 15000 });
const text = await page.locator('.xterm-rows').innerText();
if (!text.includes('GROK_TTY_BANNER')) throw new Error('terminal modal missing grok PTY banner: ' + text);
`)
	return nil
}
```