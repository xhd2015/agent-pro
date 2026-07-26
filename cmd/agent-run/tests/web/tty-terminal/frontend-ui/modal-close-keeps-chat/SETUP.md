# Scenario

**Feature**: closing terminal modal detaches browser only and keeps transcript visible

```
open terminal modal -> close -> chat transcript remains -> session status not stopped by close
```

## Preconditions

- TTY session has existing chat messages.

## Steps

1. Seed running tty session and live registry.
2. Open and close modal.
3. Verify transcript text is still visible.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "ui-modal-close"
	req.RegistryTranscript = "terminal-close-test\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByText('assistant keeps transcript').waitFor({ state: 'visible', timeout: 15000 });
await page.getByRole('button', { name: /terminal/i }).click();
await page.getByRole('dialog').waitFor({ state: 'visible', timeout: 15000 });
const close = page.getByRole('button', { name: /close|dismiss/i });
await close.click();
await page.getByRole('dialog').waitFor({ state: 'hidden', timeout: 15000 });
await page.getByText('assistant keeps transcript').waitFor({ state: 'visible', timeout: 15000 });
`)
	return nil
}
```
