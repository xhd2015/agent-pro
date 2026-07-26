# Scenario

**Feature**: terminal modal attaches to PTY and forwards keyboard input

```
terminal button click -> modal terminal surface -> PTY output visible
keyboard "modal input" + Enter -> upstream receives bytes
```

## Preconditions

- TTY session has a live registry.

## Steps

1. Seed live tty session.
2. Open modal from terminal button.
3. Type into terminal surface.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "ui-modal-attach"
	req.RegistryTranscript = "terminal-ready modal\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByRole('button', { name: /terminal/i }).click();
const dialog = page.getByRole('dialog');
await dialog.waitFor({ state: 'visible', timeout: 15000 });
await page.getByText(/terminal-ready modal|terminal/i).waitFor({ state: 'visible', timeout: 15000 });
await page.keyboard.type('modal input');
await page.keyboard.press('Enter');
await page.getByText(/modal input|echo:/i).waitFor({ state: 'visible', timeout: 15000 });
`)
	return nil
}
```
