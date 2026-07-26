# Scenario

**Feature**: terminal modal reattaches to same backend terminal after navigation

```
open modal -> close -> navigate home -> reopen same session -> open modal -> current terminal output visible
```

## Preconditions

- TTY registry remains live after browser detach.

## Steps

1. Seed running tty session and registry.
2. Open modal once.
3. Navigate home and back to same session.
4. Open modal again.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "ui-modal-reattach"
	req.RegistryTranscript = "reattach-terminal-state\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByRole('button', { name: /terminal/i }).click();
await page.getByText(/reattach-terminal-state/).waitFor({ state: 'visible', timeout: 15000 });
await page.getByRole('button', { name: /close|dismiss/i }).click();
await page.goto(`+jsQuote(req.WebBaseURL)+`, { waitUntil: 'domcontentloaded' });
await page.goto(`+jsQuote(req.WebBaseURL+"/sessions/"+req.Runner+"/"+req.SessionID)+`, { waitUntil: 'domcontentloaded' });
await page.getByRole('button', { name: /terminal/i }).click();
await page.getByText(/reattach-terminal-state/).waitFor({ state: 'visible', timeout: 15000 });
`)
	return nil
}
```
